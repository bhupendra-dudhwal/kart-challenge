package builder

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/fnv"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const cacheFile = "promo_cache.json"
const hashFile = "promo_hash.txt"

type PromoSet map[string]struct{}

func CollectGzipFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".gz") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func computeCombinedHash(files []string) (string, error) {
	h := sha256.New()

	for _, path := range files {
		f, err := os.Open(path)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(h, f)
		f.Close()
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func writeHash(h string) error {
	return os.WriteFile(hashFile, []byte(h), 0644)
}

func readStoredHash() (string, error) {
	b, err := os.ReadFile(hashFile)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

const shardCount = 256

type shard struct {
	sync.Mutex
	m map[string]int
}

var counterShards [shardCount]*shard

func init() {
	for i := range counterShards {
		counterShards[i] = &shard{m: make(map[string]int)}
	}
}

func addCode(code string) {
	h := fnv.New32a()
	h.Write([]byte(code))
	idx := h.Sum32() % shardCount
	s := counterShards[idx]

	s.Lock()
	s.m[code]++
	s.Unlock()
}

var gzipPool = sync.Pool{
	New: func() any { return new(gzip.Reader) },
}

func processFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Big buffered reader for large files
	buf := bufio.NewReaderSize(f, 4*1024*1024)

	// Reuse gzip reader from pool
	gz := gzipPool.Get().(*gzip.Reader)
	if err := gz.Reset(buf); err != nil {
		gzipPool.Put(gz)
		return err
	}
	defer func() {
		gz.Close()
		gzipPool.Put(gz)
	}()

	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)

	seen := make(map[string]struct{})

	for scanner.Scan() {
		code := scanner.Text()

		if len(code) >= 8 && len(code) <= 10 {
			seen[code] = struct{}{}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	for code := range seen {
		addCode(code)
	}

	return nil
}

func processFilesParallel(files []string) error {
	workerCount := runtime.NumCPU() * 2

	jobs := make(chan string, len(files))
	var wg sync.WaitGroup

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				if err := processFile(path); err != nil {

				}
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	wg.Wait()
	return nil
}

func saveCache(codes PromoSet) error {
	b, err := json.MarshalIndent(codes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacheFile, b, 0644)
}

func loadCache() (PromoSet, error) {
	b, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}
	var codes PromoSet
	if err := json.Unmarshal(b, &codes); err != nil {
		return nil, err
	}
	return codes, nil
}

func BootPromoSystem(files []string) (PromoSet, error) {

	if len(files) == 0 {
		return nil, errors.New("no files provided")
	}

	currentHash, err := computeCombinedHash(files)
	if err != nil {
		return nil, err
	}

	if storedHash, err := readStoredHash(); err == nil && storedHash == currentHash {
		return loadCache()
	}

	if err := processFilesParallel(files); err != nil {
		return nil, err
	}

	valid := make(PromoSet)

	for _, s := range counterShards {
		s.Lock()
		for code, count := range s.m {
			if count >= 2 {
				valid[code] = struct{}{}
			}
		}
		s.Unlock()
	}

	if err := saveCache(valid); err != nil {
		return nil, err
	}
	if err := writeHash(currentHash); err != nil {
		return nil, err
	}

	return valid, nil
}
