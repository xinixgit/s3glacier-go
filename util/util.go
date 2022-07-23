package util

import (
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

func FileLen(f *os.File) int64 {
	fi, _ := f.Stat()
	return fi.Size()
}

func Min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func CeilQuotient(a, b int64) int64 {
	return int64(math.Ceil(float64(a) / float64(b)))
}

func PanicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

// Get the range of bytes that are used in S3 requests, according to W3 rfc2616-sec14 standard.
// Note both from and to are inclusive.
func GetBytesRange(from int64, to int64) string {
	return fmt.Sprintf("bytes %d-%d/*", from, to)
}

func GetBytesRangeInt64(from int64, to int64) string {
	return fmt.Sprintf("bytes %d-%d/*", from, to)
}

func ToHexString(bytes []byte) string {
	return hex.EncodeToString(bytes)
}

func GetDBNowStr() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func ReadAllFromStream(stream io.ReadCloser) (*string, error) {
	buf := new(strings.Builder)
	if _, err := io.Copy(buf, stream); err != nil {
		return nil, err
	}

	str := buf.String()
	return &str, nil
}

func ListenForJobOutput(
	vault *string,
	jobId *string,
	onJobComplete func(jobId *string, desc *glacier.JobDescription),
	initialWait time.Duration,
	pollInterval time.Duration,
	s3glacier *glacier.Glacier) {

	fmt.Printf("Wait %ds before start polling job status.\n", int(initialWait.Seconds()))
	time.Sleep(initialWait)
	cnt := 1

	for {
		input := &glacier.DescribeJobInput{
			AccountId: aws.String("-"),
			JobId:     jobId,
			VaultName: vault,
		}
		res, err := s3glacier.DescribeJob(input)
		isCompleted := *res.Completed

		if err != nil {
			fmt.Println("Failed to pull job status from s3, ", err)
			isCompleted = false
		}

		if !isCompleted {
			fmt.Printf("Job is not ready, wait %ds before next job status poll. (%d)\n", int(pollInterval.Seconds()), cnt)
			cnt = cnt + 1

			time.Sleep(pollInterval)
			continue
		}

		onJobComplete(jobId, res)
		break
	}
}
