package util

import (
	"math"
	"os"
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
