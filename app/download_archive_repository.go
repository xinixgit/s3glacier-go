package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"s3glacier-go/domain"
	"s3glacier-go/util"
	"time"
)

type DownloadArchiveRepository interface {
	ResumeDownload(downloadID uint, ctx *DownloadJobContext)
	Download(ctx *DownloadJobContext)
}

type DownloadArchiveRepositoryImpl struct {
	svc                    domain.CloudServiceProvider
	dao                    domain.DBDAO
	jobNotificationHandler domain.JobNotificationHandler
}

func NewDownloadArchiveRepository(svc domain.CloudServiceProvider, dao domain.DBDAO, jobNotificationHandler domain.JobNotificationHandler) DownloadArchiveRepository {
	return &DownloadArchiveRepositoryImpl{
		svc:                    svc,
		dao:                    dao,
		jobNotificationHandler: jobNotificationHandler,
	}
}

type DownloadJobContext struct {
	ArchiveID       *string
	Vault           *string
	JobQueue        *string
	ChunkSize       int
	InitialWaitTime time.Duration
	WaitInterval    time.Duration
	File            *os.File
}

func (repo *DownloadArchiveRepositoryImpl) ResumeDownload(downloadID uint, ctx *DownloadJobContext) {
	download := repo.dao.GetDownloadByID(downloadID)

	jobId := download.JobId
	id := download.ID
	repo.processDownloadJob(&jobId, id, ctx)
}

func (repo DownloadArchiveRepositoryImpl) Download(ctx *DownloadJobContext) {
	jobId, err := repo.svc.InitiateArchiveRetrievalJob(ctx.ArchiveID, ctx.Vault)
	if err != nil {
		fmt.Println("Failed to initiate retrieval job")
		panic(err)
	}

	id, err := repo.insertNewDownload(jobId, ctx)
	if err != nil {
		fmt.Println("Failed to insert a new download into database.")
		panic(err)
	}

	repo.processDownloadJob(jobId, *id, ctx)
}

func (repo *DownloadArchiveRepositoryImpl) processDownloadJob(jobID *string, downloadID uint, ctx *DownloadJobContext) error {
	if _, err := repo.jobNotificationHandler.GetNotification(ctx.JobQueue, ctx.WaitInterval); err != nil {
		return err
	}

	sizeInBytes, err := repo.getDownloadableBytes(jobID, ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Now start downloading the file of %d bytes.\n", *sizeInBytes)

	if err := repo.processJobOutput(jobID, downloadID, *sizeInBytes, ctx); err != nil {
		return err
	}

	repo.updateCompletedDownload(downloadID)
	fmt.Println("Archive saved to file.")
	return nil
}

func (repo *DownloadArchiveRepositoryImpl) getDownloadableBytes(jobID *string, ctx *DownloadJobContext) (*int64, error) {
	res, err := repo.svc.DescribeJob(jobID, ctx.Vault)
	if err != nil {
		fmt.Println("Failed to pull job status from s3, ", err)
		return nil, err
	}
	if !*res.Completed {
		return nil, errors.New("received notification about job completion, but job status is not COMPLETED")
	}
	if arId := res.ArchiveId; *arId != *ctx.ArchiveID {
		return nil, fmt.Errorf("archive id not matching! Expected: %s, received: %s", *ctx.ArchiveID, *arId)
	}
	if jId := res.JobId; *jId != *jobID {
		return nil, fmt.Errorf("job id not matching! Expected: %s, received: %s", *jobID, *jId)
	}

	sizeInBytes := res.ArchiveSizeInBytes
	return sizeInBytes, nil
}

func (repo *DownloadArchiveRepositoryImpl) processJobOutput(jobId *string, downloadId uint, sizeInBytes int64, ctx *DownloadJobContext) error {
	chunkSize := int64(ctx.ChunkSize)
	rangeStart := int64(0)
	buf := make([]byte, chunkSize)

	for rangeStart < sizeInBytes {
		rangeEnd := util.Min(rangeStart+chunkSize, sizeInBytes) - 1
		bytesRange := util.GetBytesRange(rangeStart, rangeEnd)
		res, err := repo.svc.GetJobOutputByRange(jobId, &bytesRange, ctx.Vault)
		if err != nil {
			return err
		}

		read, err := io.ReadAtLeast(res.Body, buf, ctx.ChunkSize)
		if err != nil {
			return err
		}

		if read == 0 {
			break
		}

		data := buf[:read]
		expectedChecksum := *res.Checksum
		calculatedChecksum := util.ToHexString(util.ComputeSHA256TreeHashWithOneMBChunks(data))
		if expectedChecksum != calculatedChecksum {
			return fmt.Errorf("checksums are different. Expected: %s, calculated: %s", expectedChecksum, calculatedChecksum)
		}

		if _, err := ctx.File.Write(data); err != nil {
			return err
		}

		fmt.Printf("Bytes %s appended to file. ", bytesRange)
		repo.insertNewDownloadedSegment(downloadId, bytesRange)

		rangeStart = rangeEnd + 1
	}

	return nil
}
