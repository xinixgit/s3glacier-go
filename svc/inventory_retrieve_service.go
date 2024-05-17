package svc

import (
	"fmt"
	"s3glacier-go/domain"
	"time"
)

type inventoryRetrieveService interface {
	RetrieveInventory(vault *string, jobQueue *string, initialWaitTime time.Duration, waitInterval time.Duration) (*string, error)
}

type inventoryRetrieveServiceImpl struct {
	jobNotificationHandler domain.JobNotificationHandler
	csp                    domain.CloudServiceProvider
}

func NewInventoryRetrieveService(
	jobNotificationHandler domain.JobNotificationHandler,
	csp domain.CloudServiceProvider,
) inventoryRetrieveService {
	return &inventoryRetrieveServiceImpl{
		jobNotificationHandler: jobNotificationHandler,
		csp:                    csp,
	}
}

func (s *inventoryRetrieveServiceImpl) RetrieveInventory(vault *string, jobQueue *string, initialWaitTime time.Duration, waitInterval time.Duration) (*string, error) {
	if initialWaitTime > 0 {
		time.Sleep(initialWaitTime)
	}

	jobId, err := s.csp.InitiateInventoryRetrievalJob(vault)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Inventory-retrieval job started with id: %s\n", *jobId)

	if _, err = s.jobNotificationHandler.GetNotification(jobQueue, waitInterval); err != nil {
		return nil, err
	}

	output, err := s.csp.GetJobOutput(jobId, vault)
	if err != nil {
		return nil, err
	}
	return readAllFromStream(output.Body)
}
