package svc

import (
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

func fileLen(f *os.File) int64 {
	fi, _ := f.Stat()
	return fi.Size()
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func ceilQuotient(a, b int64) int64 {
	return int64(math.Ceil(float64(a) / float64(b)))
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Get the range of bytes that are used in S3 requests, according to W3 rfc2616-sec14 standard.
// Note both from and to are inclusive.
func getBytesRange(from int64, to int64) string {
	return fmt.Sprintf("bytes %d-%d/*", from, to)
}

func getBytesRangeInt64(from int64, to int64) string {
	return fmt.Sprintf("bytes %d-%d/*", from, to)
}

func toHexString(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func readAllFromStream(stream io.ReadCloser) (*string, error) {
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, stream); err != nil {
		return nil, err
	}

	str := buf.String()
	return &str, nil
}
