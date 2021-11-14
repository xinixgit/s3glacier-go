package upload

import (
	"fmt"
	"s3glacier-go/db"
	"s3glacier-go/util"

	"github.com/aws/aws-sdk-go/service/glacier"
)

func insertNewUpload(sessionId *string, filename string, vaultName *string, dao db.UploadDAO) uint {
	upload := &db.Upload{
		VaultName: *vaultName,
		Filename:  filename,
		SessionId: *sessionId,
		CreatedAt: util.GetDBNowStr(),
		Status:    db.STARTED,
	}

	if err := dao.InsertUpload(upload); err != nil {
		fmt.Printf("Insert upload failed for %s and session %s.\n", *sessionId, filename)
	}
	return upload.ID
}

func updateCompletedUpload(id uint, res *glacier.ArchiveCreationOutput, dao db.UploadDAO) {
	upload := dao.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("Failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Location = *res.Location
	upload.Checksum = *res.Checksum
	upload.ArchiveId = *res.ArchiveId
	upload.Status = db.COMPLETED
	dao.UpdateUpload(upload)
}

func updateFailedUpload(id uint, dao db.UploadDAO) {
	upload := dao.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("Failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Status = db.FAILED
	dao.UpdateUpload(upload)
}

func insertUploadedSegment(result *glacier.UploadMultipartPartOutput, segNum int, segCount int, uploadId uint, dao db.UploadDAO) {
	if uploadId == 0 {
		return
	}

	seg := &db.UploadedSegment{
		SegmentNum: segNum,
		UploadId:   uploadId,
		Checksum:   *result.Checksum,
		CreatedAt:  util.GetDBNowStr(),
	}

	if err := dao.InsertUploadedSegment(seg); err != nil {
		fmt.Printf("Insert uploaded segment failed for seg num %d, and upload id %d.\n", segNum, uploadId)
	}
}
