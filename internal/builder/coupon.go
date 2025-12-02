package builder

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func (b *appBuilder) processCouponData() error {
	start := time.Now()
	b.logger.Info("processCouponData starts")
	unzipErrs, fatalErr := b.unzip(b.config.CouponConfig.Files)
	if fatalErr != nil {
		return fatalErr
	}

	if len(unzipErrs) > 0 && !b.config.CouponConfig.IgnoreUnzipErrors {
		return fmt.Errorf("failed to unzip %d coupon files: %v", len(unzipErrs), unzipErrs)
	}

	b.logger.Info("processCouponData completes successfully", zap.Int64("duration(Micro Sec)", time.Since(start).Microseconds()))

	start = time.Now()
	b.logger.Info("buildFrequency starts")
	b.processValidData()
	b.logger.Info("buildFrequency completes successfully", zap.Int64("duration(Micro Sec)", time.Since(start).Microseconds()))

	return nil
}

func (b *appBuilder) unzip(files []string) (unzipErrs []error, executionErr error) {
	var (
		g    errgroup.Group
		mu   sync.Mutex
		errs []error
	)

	maxCpus := utils.Max(1, runtime.NumCPU()/2)

	g.SetLimit(maxCpus)

	for _, file := range files {
		f := file
		g.Go(func() error {
			if err := utils.Unzip(f, ".gz", ".txt"); err != nil {
				mu.Lock()
				errs = append(errs, fmt.Errorf("unzip failed for %s: %w", f, err))
				mu.Unlock()
			}
			return nil
		})
	}

	executionErr = g.Wait()

	return errs, executionErr
}

func (b *appBuilder) processValidData() {
	out := make(chan map[string]bool, len(b.config.CouponConfig.Files))

	for _, file := range b.config.CouponConfig.Files {
		txtFile := strings.TrimSuffix(file, ".gz") + ".txt"
		go b.parseFile(txtFile, out)
	}

	freq := make(map[string]int)
	seenFinal := make(map[string]bool)
	batch := make([]string, 0, b.config.CouponConfig.BatchSize)

	for i := 0; i < len(b.config.CouponConfig.Files); i++ {
		fileMap := <-out

		for code := range fileMap {
			if seenFinal[code] {
				continue
			}

			freq[code]++
			if freq[code] == 2 {
				batch = append(batch, code)
				seenFinal[code] = true
				delete(freq, code)
			}

			if len(batch) >= b.config.CouponConfig.BatchSize {
				b.flushToRedis(batch)
				batch = batch[:0]
			}
		}
	}

	if len(batch) > 0 {
		b.flushToRedis(batch)
	}
}

func (b *appBuilder) parseFile(filename string, out chan<- map[string]bool) {
	seen := make(map[string]bool)

	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open %s: %v", filename, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 32*1024*1024)

	for scanner.Scan() {
		code := strings.TrimSpace(scanner.Text())
		if utils.ValidateCode(code, b.config.CouponConfig.Validation) {
			seen[code] = true
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatalf("Scan error %s: %v", filename, err)
	}

	out <- seen
}

func (b *appBuilder) flushToRedis(batch []string) error {
	if len(batch) == 0 {
		return nil
	}

	args := make([]interface{}, 0, len(batch)+2)
	args = append(args, "BF.MADD", b.config.CouponConfig.BloomKey)
	for _, code := range batch {
		args = append(args, code)
	}

	if err := b.cacheRepository.Do(b.ctx, args); err != nil {
		if !b.config.CouponConfig.IgnoreUnzipErrors {
			return fmt.Errorf("redis BF.MADD error: %w", err)
		}
	}

	members := make([]interface{}, len(batch))
	for i, v := range batch {
		members[i] = v
	}

	if err := b.cacheRepository.SAdd(b.ctx, b.config.CouponConfig.ExactSet, members...); err != nil {
		if !b.config.CouponConfig.IgnoreUnzipErrors {
			return fmt.Errorf("redis BF.MADD error: %w", err)
		}
	}
	return nil
}
