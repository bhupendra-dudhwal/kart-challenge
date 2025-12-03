package builder

import (
	"bufio"
	"fmt"
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
		if !b.config.CouponConfig.IgnoreUnzipErrors {
			b.logger.Warn("Unzip execution error ignored", zap.Error(fatalErr))
			return nil
		}
		return fatalErr
	}

	if len(unzipErrs) > 0 {
		if b.config.CouponConfig.IgnoreUnzipErrors {
			b.logger.Warn("Unzip error ignored", zap.Errors("unzipErrs", unzipErrs))
			return nil
		}
		return fmt.Errorf("failed to unzip %d coupon files: %v", len(unzipErrs), unzipErrs)
	}

	b.logger.Info("processCouponData unzip completes", zap.Int64("duration(Micro)", time.Since(start).Microseconds()))

	start = time.Now()
	b.logger.Info("buildFrequency starts")
	if err := b.processValidData(); err != nil {
		if b.config.CouponConfig.IgnoreUnzipErrors {
			b.logger.Warn("Processing coupon data error ignored", zap.Error(err))
			return nil
		}
		return fmt.Errorf("failed to process coupon data: %w", err)
	}
	b.logger.Info("buildFrequency completes", zap.Int64("duration(ms)", time.Since(start).Milliseconds()))

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

func (b *appBuilder) processValidData() error {
	out := make(chan map[string]bool, len(b.config.CouponConfig.Files))
	errCh := make(chan error, len(b.config.CouponConfig.Files))

	for _, file := range b.config.CouponConfig.Files {
		txtFile := strings.TrimSuffix(file, ".gz") + ".txt"
		go func(f string) {
			if err := b.parseFile(f, out); err != nil {
				errCh <- fmt.Errorf("parse error for file %s: %w", f, err)
			} else {
				errCh <- nil
			}
		}(txtFile)
	}

	freq := make(map[string]int)
	seenFinal := make(map[string]bool)
	batch := make([]string, 0, b.config.CouponConfig.BatchSize)

	for i := 0; i < len(b.config.CouponConfig.Files); i++ {
		select {
		case fileMap := <-out:
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
					if err := b.flushToRedis(batch); err != nil {
						return err
					}
					batch = batch[:0]
				}
			}
		case err := <-errCh:
			if err != nil && !b.config.CouponConfig.IgnoreUnzipErrors {
				return err
			} else if err != nil {
				b.logger.Warn("Parse file error ignored", zap.Error(err))
			}
		}
	}

	if len(batch) > 0 {
		if err := b.flushToRedis(batch); err != nil {
			return err
		}
	}

	return nil
}

func (b *appBuilder) parseFile(filename string, out chan<- map[string]bool) error {
	seen := make(map[string]bool)

	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", filename, err)
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
		return fmt.Errorf("scan error in file %s: %w", filename, err)
	}

	out <- seen
	return nil
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
		b.logger.Info("bloom filter creation", zap.Error(err))
		if !strings.Contains(err.Error(), "item exists") {
			return fmt.Errorf("failed to create bloom filter: %w", err)
		}
	}

	members := make([]interface{}, len(batch))
	for i, code := range batch {
		members[i] = code
	}

	fmt.Println("\n\n\n ")
	b.logger.Info("members", zap.Any("members", members))
	fmt.Println("\n\n\n ")

	if err := b.cacheRepository.SAdd(b.ctx, b.config.CouponConfig.ExactSet, members...); err != nil {
		return fmt.Errorf("redis SADD error: %w", err)
	}

	return nil
}
