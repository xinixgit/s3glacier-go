package util

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"time"
)

func FileLen(f *os.File) int64 {
	fi, _ := f.Stat()
	return fi.Size()
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CeilQuotient(a, b int) int {
	return int(math.Ceil(float64(a) / float64(b)))
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Get the range of bytes that are used in S3 requests, according to W3 rfc2616-sec14 standard.
// Note both from and to are inclusive.
func GetBytesRange(from int, to int) string {
	return fmt.Sprintf("bytes %d-%d/*", from, to)
}

func ToHexString(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func GetDBNowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}
