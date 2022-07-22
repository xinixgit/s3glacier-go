package app

import (
	"fmt"
	"s3glacier-go/domain"
	"time"
)

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
	expiredUploads, err := p.dao.GetExpiredUpload(vault)
	if err != nil {
		return err
	}

	if len(expiredUploads) == 0 {
		fmt.Println("No expired archive found.")
		return nil
	}

	for _, upload := range expiredUploads {
		id := upload.ArchiveId
		if err := p.svc.DeleteArchive(&id, vault); err != nil {
			fmt.Printf("Fail to delete archive: %s\n", id)
			return err
		}

		fmt.Printf("Mark archive as deleted: %d\n", upload.ID)

		upload.Status = domain.DELETED
		p.dao.UpdateUpload(&upload)
	}

	return nil
}
