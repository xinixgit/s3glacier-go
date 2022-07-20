package app

import (
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
	svc domain.CloudServiceProvider
	dao domain.DBDAO
}

func NewDownloadArchiveRepository(svc domain.CloudServiceProvider, dao domain.DBDAO) DownloadArchiveRepository {
	return &DownloadArchiveRepositoryImpl{
		svc: svc,
		dao: dao,
	}
}

type DownloadJobContext struct {
	ArchiveID       *string
	Vault           *string
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

func (repo *DownloadArchiveRepositoryImpl) processDownloadJob(jobId *string, downloadID uint, ctx *DownloadJobContext) {
	repo.svc.OnJobComplete(jobId, ctx.ArchiveID, ctx.Vault, ctx.WaitInterval, func(sizeInBytes int) {
		repo.processJobOutput(jobId, downloadID, sizeInBytes, ctx)
		repo.updateCompletedDownload(downloadID)
		fmt.Println("Archive saved to file.")
	})
}

func (repo *DownloadArchiveRepositoryImpl) processJobOutput(jobId *string, downloadId uint, sizeInBytes int, ctx *DownloadJobContext) {
	chunkSize := ctx.ChunkSize
	rangeStart := 0
	buf := make([]byte, chunkSize)

	for rangeStart < sizeInBytes {
		rangeEnd := util.Min(rangeStart+chunkSize, sizeInBytes) - 1
		bytesRange := util.GetBytesRange(rangeStart, rangeEnd)
		res, err := repo.svc.GetJobOutputByRange(jobId, &bytesRange, ctx.Vault)
		if err != nil {
			panic(err)
		}

		read, err := io.ReadAtLeast(res.Body, buf, ctx.ChunkSize)
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

		if _, err := ctx.File.Write(data); err != nil {
			fmt.Printf("Writing data to file failed for byte range %s.\n", bytesRange)
			panic(err)
		}

		fmt.Printf("Bytes %s appended to file. ", bytesRange)
		repo.insertNewDownloadedSegment(downloadId, bytesRange)

		rangeStart = rangeEnd + 1
	}
}
