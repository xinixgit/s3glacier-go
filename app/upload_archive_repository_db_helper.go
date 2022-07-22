package app

import (
	"fmt"
	"s3glacier-go/domain"
	"s3glacier-go/util"
)

func (repo *UploadArchiveRepositoryImpl) insertNewUpload(sessionId *string, filename string, vaultName *string) uint {
	upload := &domain.Upload{
		VaultName: *vaultName,
		Filename:  filename,
		SessionId: *sessionId,
		CreatedAt: util.GetDBNowStr(),
		Status:    domain.STARTED,
	}

	if err := repo.dao.InsertUpload(upload); err != nil {
		fmt.Printf("Insert upload failed for %s and session %s.\n", *sessionId, filename)
	}
	return upload.ID
}

func (repo *UploadArchiveRepositoryImpl) updateCompletedUpload(id uint, res *domain.ArchiveCreationOutput) {
	upload := repo.dao.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("Failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Location = *res.Location
	upload.Checksum = *res.Checksum
	upload.ArchiveId = *res.ArchiveId
	upload.Status = domain.COMPLETED
	repo.dao.UpdateUpload(upload)
}

func (repo *UploadArchiveRepositoryImpl) updateFailedUpload(id uint) {
	upload := repo.dao.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("Failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Status = domain.FAILED
	repo.dao.UpdateUpload(upload)
}

func (repo *UploadArchiveRepositoryImpl) insertUploadedSegment(checksum *string, segNum int, segCount int, uploadId uint) {
	if uploadId == 0 {
		return
	}

	seg := &domain.UploadedSegment{
		SegmentNum: segNum,
		UploadId:   uploadId,
		Checksum:   *checksum,
		CreatedAt:  util.GetDBNowStr(),
	}

	if err := repo.dao.InsertUploadedSegment(seg); err != nil {
		fmt.Printf("Insert uploaded segment failed for seg num %d, and upload id %d.\n", segNum, uploadId)
	}
}
