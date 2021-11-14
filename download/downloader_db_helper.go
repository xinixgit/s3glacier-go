package download

import (
	"fmt"
	"s3glacier-go/db"
	"s3glacier-go/util"
)

func insertNewDownload(jobId *string, archiveId *string, outputFile *string, vaultName *string, dao db.DownloadDAO) (id uint) {
	dl := &db.Download{
		VaultName: *vaultName,
		JobId:     *jobId,
		ArchiveId: *archiveId,
		CreatedAt: util.GetDBNowStr(),
		Status:    db.STARTED,
	}
	if err := dao.InsertDownload(dl); err != nil {
		panic(err)
	}

	id = dl.ID
	return
}

func updateCompletedDownload(id uint, dao db.DownloadDAO) {
	dl := dao.GetDownloadByID(id)
	dl.Status = db.COMPLETED
	dao.UpdateDownload(dl)
}

func insertNewDownloadedSegment(dlID uint, bytesRange string, dao db.DownloadDAO) {
	seg := &db.DownloadedSegment{
		DownloadId: dlID,
		BytesRange: bytesRange,
		CreatedAt:  util.GetDBNowStr(),
	}
	if err := dao.InsertDownloadedSegment(seg); err != nil {
		fmt.Printf("Failed to insert bytes %s into database for download id %d", bytesRange, dlID)
		return
	}

	fmt.Printf("Seg %d saved to disk for dl %d\n", seg.ID, dlID)
}
