package upload

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/glacier"
	"time"
	"xddd/s3glacier/db"
)

func insertNewUpload(sessionId *string, filename string, u *S3GlacierUploader) uint {
	upload := &db.Upload{
		VaultName: *u.Vault,
		Filename:  filename,
		SessionId: *sessionId,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
		Status:    db.STARTED,
	}

	err := u.DBDAO.InsertUpload(upload)
	if err != nil {
		fmt.Printf("Insert upload failed for %s and session %s.\n", *sessionId, filename)
	}
	return upload.ID
}

func updateCompletedUpload(id uint, res *glacier.ArchiveCreationOutput, u *S3GlacierUploader) {
	upload := u.DBDAO.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("Failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Location = *res.Location
	upload.Checksum = *res.Checksum
	upload.ArchiveId = *res.ArchiveId
	upload.Status = db.COMPLETED
	u.DBDAO.UpdateUpload(upload)
}

func updateFailedUpload(id uint, u *S3GlacierUploader) {
	upload := u.DBDAO.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("Failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Status = db.FAILED
	u.DBDAO.UpdateUpload(upload)
}

func insertUploadedSegment(result *glacier.UploadMultipartPartOutput, segNum int, segCount int, uploadId uint, u *S3GlacierUploader) {
	if uploadId == 0 {
		return
	}

	seg := &db.UploadedSegment{
		SegmentNum: segNum,
		UploadId:   uploadId,
		Checksum:   *result.Checksum,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
	}

	err := u.DBDAO.InsertUploadedSegment(seg)
	if err != nil {
		fmt.Printf("Insert uploaded segment failed for seg num %d, and upload id %d.\n", segNum, uploadId)
	}
}
