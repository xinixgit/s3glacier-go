package app

import (
	"fmt"
	"s3glacier-go/domain"
	"time"
)

const DefaultAccountID = "-"

const TIMESTAMP_LAYOUT = "2006-01-02T15:04:05Z"

const MIN_HOLDING_DURATION = 91 * time.Hour * 24

type DeleteArchiveRepository struct {
	dao domain.DBDAO
	svc domain.CloudServiceProvider
}

func NewDeleteArchiveRepository(dao domain.DBDAO, svc domain.CloudServiceProvider) *DeleteArchiveRepository {
	return &DeleteArchiveRepository{
		dao: dao,
		svc: svc,
	}
}

func (p *DeleteArchiveRepository) DeleteExpiredArchive(vault *string) error {
	archiveIds := []string{}
	expiredUploads, err := p.dao.GetExpiredUpload(vault)
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
		if err := p.deleteArchive(&id, vault); err != nil {
			fmt.Printf("Fail to delete archive: %s\n", id)
			return err
		}

		fmt.Printf("Deleted archive: %s\n", id)
	}

	return nil
}

func (p *DeleteArchiveRepository) deleteArchive(archiveID *string, vault *string) error {
	accountID := DefaultAccountID
	return p.svc.DeleteArchive(&accountID, archiveID, vault)
}
