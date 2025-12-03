package utils

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/bhupendra-dudhwal/kart-challenge/internal/constants"
	"github.com/bhupendra-dudhwal/kart-challenge/internal/core/models"

	"github.com/valyala/fasthttp"
)

func FileExists(path string) error {
	info, err := os.Stat(path)
	if err == nil {
		if info.IsDir() {
			return fmt.Errorf("path exists but is a directory: %s", path)
		}
		return nil
	}

	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", path)
	}

	return fmt.Errorf("failed to check file: %w", err)
}

func CtxValue[T any](ctx context.Context, key constants.CtxKey) (T, bool) {
	v := ctx.Value(key)
	return ctxValue[T](v)
}

func HeaderValue[T any](ctx *fasthttp.RequestCtx, key string) (T, bool) {
	raw := ctx.Request.Header.Peek(key)
	if raw == nil {
		var zero T
		return zero, false
	}
	return convert[T](raw)
}

func ctxValue[T any](v any) (T, bool) {
	var zero T
	if v == nil {
		return zero, false
	}

	switch any(zero).(type) {
	case string:
		strVal, ok := v.(string)
		if !ok {
			return zero, false
		}
		return any(strVal).(T), true
	case int64:
		var strVal string
		switch t := v.(type) {
		case string:
			strVal = t
		case []byte:
			strVal = string(t)
		default:
			return zero, false
		}
		num, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return zero, false
		}
		return any(num).(T), true
	case int:
		var strVal string
		switch t := v.(type) {
		case string:
			strVal = t
		case []byte:
			strVal = string(t)
		default:
			return zero, false
		}
		num, err := strconv.Atoi(strVal)
		if err != nil {
			return zero, false
		}
		return any(num).(T), true
	default:
		return zero, false
	}
}

func PathParamValue[T any](ctx *fasthttp.RequestCtx, key string) (T, bool) {
	v := ctx.UserValue(key)
	return ctxValue[T](v)
}

func convert[T any](v []byte) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case string:
		return any(string(v)).(T), true

	case []byte:
		return any(v).(T), true

	default:
		return zero, false
	}
}

var gzipReaderPool = sync.Pool{
	New: func() any { return new(gzip.Reader) },
}

func Unzip(gzFile, sourceExt, destExt string) error {
	destFile := strings.TrimSuffix(gzFile, sourceExt) + destExt

	if _, err := os.Stat(destFile); err == nil {
		return nil
	}

	tmp := destFile + ".tmp"

	if err := unzipGZ(gzFile, tmp); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to unzip %s: %w", gzFile, err)
	}

	return os.Rename(tmp, destFile)
}

func unzipGZ(src, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer f.Close()

	gz := gzipReaderPool.Get().(*gzip.Reader)
	if err := gz.Reset(f); err != nil {
		gzipReaderPool.Put(gz)
		return fmt.Errorf("gzip reset: %w", err)
	}
	defer func() {
		gz.Close()
		gzipReaderPool.Put(gz)
	}()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer out.Close()

	buf := make([]byte, 1*1024*1024)

	_, err = io.CopyBuffer(out, gz, buf)
	return err
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func isUpper(ch byte) bool { return ch >= 'A' && ch <= 'Z' }
func isLower(ch byte) bool { return ch >= 'a' && ch <= 'z' }
func isDigit(ch byte) bool { return ch >= '0' && ch <= '9' }

func ValidateCode(code string, cfg *models.CouponValidator) bool {
	ln := len(code)

	if ln < cfg.MinLength || ln > cfg.MaxLength {
		return false
	}

	for i := 0; i < ln; i++ {
		ch := code[i]

		switch cfg.AllowedCharacters {

		case constants.Alphanumeric:
			if !(isUpper(ch) || isLower(ch) || isDigit(ch)) {
				return false
			}

		case constants.Digits:
			if !isDigit(ch) {
				return false
			}

		case constants.Letters:
			if !(isUpper(ch) || isLower(ch)) {
				return false
			}

		case constants.Uppercase:
			if !isUpper(ch) {
				return false
			}

		case constants.Lowercase:
			if !isLower(ch) {
				return false
			}

		default:
			return false
		}
	}

	return true
}

func Sanitize(str string) string {
	clean := strings.TrimSpace(str)
	return strings.ToValidUTF8(clean, "")
}

func RoundFloat64(value float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(value*factor) / factor
}
