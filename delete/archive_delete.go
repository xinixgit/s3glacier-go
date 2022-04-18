package delete

import (
	"fmt"
	"s3glacier-go/db"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

type ArchiveDeleteHandler struct {
	Vault     *string
	UlDAO     db.UploadDAO
	S3glacier *glacier.Glacier
}

type Archive struct {
	ArchiveId          string
	ArchiveDescription string
	CreationDate       string
	Size               int64
	SHA256TreeHash     string
}

type Inventory struct {
	VaultARN      string
	InventoryDate string
	ArchiveList   []Archive
}

const TIMESTAMP_LAYOUT = "2006-01-02T15:04:05Z"

const MIN_HOLDING_DURATION = 91 * time.Hour * 24

func (h ArchiveDeleteHandler) DeleteExpired() error {
	archiveIds := []string{}
	expiredUploads, err := h.UlDAO.GetExpiredUpload(h.Vault)
	if err != nil {
		return err
	}

	for _, upload := range expiredUploads {
		archiveIds = append(archiveIds, upload.ArchiveId)
	}

	if len(archiveIds) == 0 {
		fmt.Println("No expired archive found.")
		return nil
	}

	for _, id := range archiveIds {
		if err := h.deleteExpired(&id); err != nil {
			fmt.Printf("Fail to delete archive: %s\n", id)
			return err
		}

		fmt.Printf("Deleted archive: %s\n", id)
	}

	return nil
}

func (h ArchiveDeleteHandler) deleteExpired(archiveId *string) error {
	input := &glacier.DeleteArchiveInput{
		AccountId: aws.String("-"),
		ArchiveId: archiveId,
		VaultName: h.Vault,
	}

	_, err := h.S3glacier.DeleteArchive(input)
	return err
}
