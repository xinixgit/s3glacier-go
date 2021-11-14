package download

import (
	"fmt"
	"io"
	"os"
	"s3glacier-go/db"
	"s3glacier-go/util"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

const TEN_MINUTE = 10 * time.Minute

type S3GlacierDownloader struct {
	Vault           *string
	ChunkSize       int
	InitialWaitTime time.Duration
	S3glacier       *glacier.Glacier
	DlDAO           db.DownloadDAO
}

func (dler S3GlacierDownloader) ResumeDownload(dl *db.Download, archiveId *string, f *os.File) {
	jobId := dl.JobId
	id := dl.ID

	processDownloadJob(&jobId, archiveId, id, f, dler)
}

func (dler S3GlacierDownloader) Download(archiveId *string, f *os.File) {
	filename := f.Name()
	jobId := dler.initiateDownloadJob(archiveId)
	id := insertNewDownload(jobId, archiveId, &filename, dler.Vault, dler.DlDAO)

	processDownloadJob(jobId, archiveId, id, f, dler)
}

func processDownloadJob(jobId *string, archiveId *string, uploadId uint, f *os.File, dler S3GlacierDownloader) {
	dler.listenForJobOutput(jobId, archiveId, func(sizeInBytes int) {
		dler.processJobOutput(jobId, uploadId, sizeInBytes, f)
		updateCompletedDownload(uploadId, dler.DlDAO)
		fmt.Println("Archive saved to file.")
	})
}

func (dler S3GlacierDownloader) initiateDownloadJob(archiveId *string) *string {
	input := &glacier.InitiateJobInput{
		AccountId: aws.String("-"),
		JobParameters: &glacier.JobParameters{
			ArchiveId: archiveId,
			Type:      aws.String("archive-retrieval"),
		},
		VaultName: dler.Vault,
	}

	res, err := dler.S3glacier.InitiateJob(input)
	if err != nil {
		panic(err)
	}
	return res.JobId
}

// To listen for a job's completion event, we periodically pull the job status from s3, after
// sleeping `initialWaitTimeInHours` hours since the jobs are not going to be ready during this
// time anyway. We could alternatively use SNS + SQS, but that involves added complexities.
func (dler S3GlacierDownloader) listenForJobOutput(jobId *string, archiveId *string, callback func(int)) {
	fmt.Printf("Wait %ds before start polling job status.\n", int(dler.InitialWaitTime.Seconds()))
	time.Sleep(dler.InitialWaitTime)

	for {
		fmt.Printf("Wait %ds before next job status poll.\n", int(TEN_MINUTE.Seconds()))
		time.Sleep(TEN_MINUTE)

		input := &glacier.DescribeJobInput{
			AccountId: aws.String("-"),
			JobId:     jobId,
			VaultName: dler.Vault,
		}
		res, err := dler.S3glacier.DescribeJob(input)

		if err != nil {
			fmt.Println("Failed to pull job status from s3, ", err)
			continue
		}
		if !*res.Completed {
			fmt.Println("Job is not ready.")
			continue
		}
		if arId := res.ArchiveId; *arId != *archiveId {
			msg := fmt.Sprintf("Archive id not matching! Expected: %s, received: %s", *archiveId, *arId)
			panic(msg)
		}
		if jId := res.JobId; *jId != *jobId {
			msg := fmt.Sprintf("Job id not matching! Expected: %s, received: %s", *jobId, *jId)
			panic(msg)
		}

		sizeInBytes := res.ArchiveSizeInBytes
		fmt.Printf("Now start downloading the file of %d bytes.\n", *sizeInBytes)
		callback(int(*sizeInBytes))

		break
	}
}

func (dler S3GlacierDownloader) processJobOutput(jobId *string, downloadId uint, sizeInBytes int, f *os.File) {
	rangeStart := 0
	buf := make([]byte, dler.ChunkSize)

	for rangeStart < sizeInBytes {
		rangeEnd := util.Min(rangeStart+dler.ChunkSize, sizeInBytes) - 1
		bytesRange := util.GetBytesRange(rangeStart, rangeEnd)
		input := &glacier.GetJobOutputInput{
			AccountId: aws.String("-"),
			JobId:     jobId,
			Range:     aws.String(bytesRange),
			VaultName: dler.Vault,
		}

		res, err := dler.S3glacier.GetJobOutput(input)
		if err != nil {
			panic(err)
		}

		read, err := io.ReadAtLeast(res.Body, buf, dler.ChunkSize)
		if read == 0 {
			break
		}

		data := buf[:read]
		expectedChecksum := *res.Checksum
		calculatedChecksum := util.ToHexString(util.ComputeSHA256TreeHashWithOneMBChunks(data))
		if expectedChecksum != calculatedChecksum {
			msg := fmt.Sprintf("Checksums are different. Expected: %s, calculated: %s.", expectedChecksum, calculatedChecksum)
			panic(msg)
		}

		if _, err := f.Write(data); err != nil {
			fmt.Printf("Writing data to file failed for byte range %s.\n", bytesRange)
			panic(err)
		}

		fmt.Printf("Bytes %s appended to file. ", bytesRange)
		insertNewDownloadedSegment(downloadId, bytesRange, dler.DlDAO)

		rangeStart = rangeEnd + 1
	}
}
