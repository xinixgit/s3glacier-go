package app

import (
	"fmt"
	"s3glacier-go/domain"
	"s3glacier-go/util"
)

func (repo *DownloadArchiveRepositoryImpl) insertNewDownload(jobId *string, ctx *DownloadJobContext) (*uint, error) {
	download := &domain.Download{
		VaultName: *ctx.Vault,
		JobId:     *jobId,
		ArchiveId: *ctx.ArchiveID,
		CreatedAt: util.GetDBNowStr(),
		Status:    domain.STARTED,
	}
	if err := repo.dao.InsertDownload(download); err != nil {
		return nil, err
	}

	return &download.ID, nil
}

func (repo *DownloadArchiveRepositoryImpl) updateCompletedDownload(id uint) {
	download := repo.dao.GetDownloadByID(id)
	download.Status = domain.COMPLETED
	repo.dao.UpdateDownload(download)
}

func (repo *DownloadArchiveRepositoryImpl) insertNewDownloadedSegment(downloadID uint, bytesRange string) {
	seg := &domain.DownloadedSegment{
		DownloadId: downloadID,
		BytesRange: bytesRange,
		CreatedAt:  util.GetDBNowStr(),
	}
	if err := repo.dao.InsertDownloadedSegment(seg); err != nil {
		fmt.Printf("Failed to insert bytes %s into database for download id %d", bytesRange, downloadID)
		return
	}

	fmt.Printf("Seg %d saved to disk for dl %d\n", seg.ID, downloadID)
}
