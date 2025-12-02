package builder

import (
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"runtime"
	"sync"
)

const cacheFile = "promo_cache.json"
const hashFile = "promo_hash.txt"

type PromoSet map[string]struct{}

type fileTask struct {
	path string
}

type fileResult struct {
	codes map[string]struct{} // unique codes in this file
	err   error
}

//
// ---------- FILE PROCESSING (Parallel) -----------
//

// Reads a single file and returns unique valid promo codes found in that file
func processFile(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	scanner := bufio.NewScanner(gz)
	scanner.Buffer(make([]byte, 32*1024), 1024*1024)

	seen := make(map[string]struct{}) // unique codes per file

	for scanner.Scan() {
		code := scanner.Text()
		if len(code) >= 8 && len(code) <= 10 {
			seen[code] = struct{}{}
		}
	}

	return seen, scanner.Err()
}

// Runs file processing in a worker pool (parallel)
func processFilesParallel(files []string) map[string]int {
	workerCount := runtime.NumCPU() * 2

	jobs := make(chan fileTask, len(files))
	results := make(chan fileResult, len(files))

	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range jobs {
				data, err := processFile(task.path)
				results <- fileResult{codes: data, err: err}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, f := range files {
			jobs <- fileTask{path: f}
		}
		close(jobs)
	}()

	// Close results after work finish
	go func() {
		wg.Wait()
		close(results)
	}()

	counts := make(map[string]int)

	// Merge results
	for res := range results {
		if res.err != nil {
			continue
		}
		for code := range res.codes {
			counts[code]++
		}
	}

	return counts
}

//
// ---------- HASH SYSTEM (Parallel) -----------
//

// Hash single file
func computeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Parallel file hashing
func computeCombinedHash(files []string) (string, error) {
	type hashRes struct {
		h   string
		err error
	}

	var wg sync.WaitGroup
	out := make(chan hashRes, len(files))

	for _, f := range files {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			h, err := computeFileHash(file)
			out <- hashRes{h: h, err: err}
		}(f)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	h := sha256.New()

	for res := range out {
		if res.err != nil {
			return "", res.err
		}
		h.Write([]byte(res.h))
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

//
// ---------- CACHE SYSTEM -----------
//

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

//
// ---------- MAIN BOOT LOGIC -----------
//

// BootPromoSystem:
// 1. Computes combined file hash (parallel)
// 2. If unchanged → load cache
// 3. Else read all files (parallel worker pool)
// 4. Save valid promo codes where code appears in at least TWO FILES
func BootPromoSystem(files []string) (PromoSet, error) {

	currentHash, err := computeCombinedHash(files)
	if err != nil {
		return nil, err
	}

	storedHash, err := readStoredHash()
	if err == nil && storedHash == currentHash {
		return loadCache()
	}

	// Process files in parallel
	countMap := processFilesParallel(files)

	valid := make(PromoSet)
	for code, count := range countMap {
		if count >= 2 { // appears in at least 2 different files
			valid[code] = struct{}{}
		}
	}

	saveCache(valid)
	writeHash(currentHash)

	return valid, nil
}
