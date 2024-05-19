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
	notificationHandler domain.NotificationHandler
	csp                 domain.CloudServiceProvider
}

func NewInventoryRetrieveService(
	notificationHandler domain.NotificationHandler,
	csp domain.CloudServiceProvider,
) inventoryRetrieveService {
	return &inventoryRetrieveServiceImpl{
		notificationHandler: notificationHandler,
		csp:                 csp,
	}
}

func (s *inventoryRetrieveServiceImpl) RetrieveInventory(
	vault *string,
	jobQueue *string,
	initialWaitTime time.Duration,
	waitInterval time.Duration,
) (*string, error) {
	jobId, err := s.csp.InitiateInventoryRetrievalJob(vault)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Inventory-retrieval job started with id: %s\n", *jobId)

	if initialWaitTime > 0 {
		time.Sleep(initialWaitTime)
	}

	// wait for job's completion via notifications
	_, err = s.notificationHandler.PollWithInterval(jobQueue, waitInterval)
	if err != nil {
		return nil, err
	}

	output, err := s.csp.GetJobOutput(jobId, vault)
	if err != nil {
		return nil, err
	}
	return ReadAllFromStream(output.Body)
}
