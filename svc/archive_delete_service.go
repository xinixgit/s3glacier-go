package svc

import (
	"fmt"
	"s3glacier-go/domain"
	"time"
)

const TIMESTAMP_LAYOUT = "2006-01-02T15:04:05Z"

const MIN_HOLDING_DURATION = 91 * time.Hour * 24

type archiveDeleteService struct {
	dao domain.DBDAO
	csp domain.CloudServiceProvider
}

func NewArchiveDeleteService(dao domain.DBDAO, csp domain.CloudServiceProvider) *archiveDeleteService {
	return &archiveDeleteService{
		dao: dao,
		csp: csp,
	}
}

func (s *archiveDeleteService) DeleteExpiredArchive(vault *string) error {
	expiredUploads, err := s.dao.GetExpiredUpload(vault)
	if err != nil {
		return err
	}

	if len(expiredUploads) == 0 {
		fmt.Println("No expired archive found.")
		return nil
	}

	for _, upload := range expiredUploads {
		id := upload.ArchiveId
		if err := s.csp.DeleteArchive(&id, vault); err != nil {
			return fmt.Errorf("cloud service provider failed to delete archive %s: %w", id, err)
		}

		upload.Status = domain.DELETED
		if err := s.dao.UpdateUpload(&upload); err != nil {
			return fmt.Errorf("unable to mark upload %d as deleted: %w", upload.ID, err)
		}

		fmt.Printf("Upload %d is deleted\n", upload.ID)
	}

	return nil
}
