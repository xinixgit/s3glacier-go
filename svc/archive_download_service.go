package svc

import (
	"errors"
	"fmt"
	"io"
	"os"
	"s3glacier-go/domain"
	"s3glacier-go/util"
	"time"
)

type DownloadJobContext struct {
	ArchiveID       *string
	Vault           *string
	JobQueue        *string
	ChunkSize       int
	InitialWaitTime time.Duration
	WaitInterval    time.Duration
	File            *os.File
}

type archiveDownloadService interface {
	ResumeDownload(downloadID uint, ctx *DownloadJobContext)
	Download(ctx *DownloadJobContext)
}

type archiveDownloadServiceImpl struct {
	csp                 domain.CloudServiceProvider
	dao                 domain.DBDAO
	notificationHandler domain.NotificationHandler
}

func NewArchiveDownloadService(csp domain.CloudServiceProvider, dao domain.DBDAO, jobNotificationHandler domain.NotificationHandler) archiveDownloadService {
	return &archiveDownloadServiceImpl{
		csp:                 csp,
		dao:                 dao,
		notificationHandler: jobNotificationHandler,
	}
}

func (s *archiveDownloadServiceImpl) ResumeDownload(downloadID uint, ctx *DownloadJobContext) {
	download := s.dao.GetDownloadByID(downloadID)

	jobId := download.JobId
	id := download.ID
	s.processDownloadJob(&jobId, id, ctx)
}

func (s archiveDownloadServiceImpl) Download(ctx *DownloadJobContext) {
	jobId, err := s.csp.InitiateArchiveRetrievalJob(ctx.ArchiveID, ctx.Vault)
	if err != nil {
		fmt.Println("Failed to initiate retrieval job")
		panic(err)
	}

	id, err := s.insertNewDownload(jobId, ctx)
	if err != nil {
		fmt.Println("Failed to insert a new download into database.")
		panic(err)
	}

	s.processDownloadJob(jobId, *id, ctx)
}

func (s *archiveDownloadServiceImpl) processDownloadJob(jobID *string, downloadID uint, ctx *DownloadJobContext) error {
	// wait for job's completion via notifications
	_, err := s.notificationHandler.PollWithInterval(ctx.JobQueue, ctx.WaitInterval)
	if err != nil {
		return err
	}

	sizeInBytes, err := s.getDownloadableBytes(jobID, ctx)
	if err != nil {
		return err
	}
	fmt.Printf("Now start downloading the file of %d bytes.\n", *sizeInBytes)

	if err := s.processJobOutput(jobID, downloadID, *sizeInBytes, ctx); err != nil {
		return err
	}

	s.updateCompletedDownload(downloadID)
	fmt.Println("Archive saved to file.")
	return nil
}

func (s *archiveDownloadServiceImpl) getDownloadableBytes(jobID *string, ctx *DownloadJobContext) (*int64, error) {
	res, err := s.csp.DescribeJob(jobID, ctx.Vault)
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

func (s *archiveDownloadServiceImpl) processJobOutput(jobId *string, downloadId uint, sizeInBytes int64, ctx *DownloadJobContext) error {
	chunkSize := int64(ctx.ChunkSize)
	rangeStart := int64(0)
	buf := make([]byte, chunkSize)

	for rangeStart < sizeInBytes {
		rangeEnd := min(rangeStart+chunkSize, sizeInBytes) - 1
		bytesRange := getBytesRange(rangeStart, rangeEnd)
		res, err := s.csp.GetJobOutputByRange(jobId, &bytesRange, ctx.Vault)
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
		calculatedChecksum := toHexString(util.ComputeSHA256TreeHashWithOneMBChunks(data))
		if expectedChecksum != calculatedChecksum {
			return fmt.Errorf("checksums are different. Expected: %s, calculated: %s", expectedChecksum, calculatedChecksum)
		}

		if _, err := ctx.File.Write(data); err != nil {
			return err
		}

		fmt.Printf("Bytes %s appended to file. ", bytesRange)
		s.insertNewDownloadedSegment(downloadId, bytesRange)

		rangeStart = rangeEnd + 1
	}

	return nil
}

func (s *archiveDownloadServiceImpl) insertNewDownload(jobId *string, ctx *DownloadJobContext) (*uint, error) {
	download := &domain.Download{
		VaultName: *ctx.Vault,
		JobId:     *jobId,
		ArchiveId: *ctx.ArchiveID,
		CreatedAt: util.GetDBNowStr(),
		Status:    domain.STARTED,
	}
	if err := s.dao.InsertDownload(download); err != nil {
		return nil, err
	}

	return &download.ID, nil
}

func (s *archiveDownloadServiceImpl) updateCompletedDownload(id uint) {
	download := s.dao.GetDownloadByID(id)
	download.Status = domain.COMPLETED
	s.dao.UpdateDownload(download)
}

func (s *archiveDownloadServiceImpl) insertNewDownloadedSegment(downloadID uint, bytesRange string) {
	seg := &domain.DownloadedSegment{
		DownloadId: downloadID,
		BytesRange: bytesRange,
		CreatedAt:  util.GetDBNowStr(),
	}
	if err := s.dao.InsertDownloadedSegment(seg); err != nil {
		fmt.Printf("Failed to insert bytes %s into database for download id %d", bytesRange, downloadID)
		return
	}

	fmt.Printf("Seg %d saved to disk for dl %d\n", seg.ID, downloadID)
}
