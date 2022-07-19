package app

import (
	"fmt"
	"s3glacier-go/domain"
	"time"
)

type InventoryRetrievalRepository interface {
	RetrieveInventory(vault *string, jobQueue *string, initialWaitTime time.Duration, waitInterval time.Duration) (*string, error)
}

type InventoryRetrievalRepositoryImpl struct {
	jobNotificationHandler domain.JobNotificationHandler
	svc                    domain.CloudServiceProvider
}

func NewInventoryRetrievalRepository(
	jobNotificationHandler domain.JobNotificationHandler,
	svc domain.CloudServiceProvider,
) InventoryRetrievalRepository {
	return &InventoryRetrievalRepositoryImpl{
		jobNotificationHandler: jobNotificationHandler,
		svc:                    svc,
	}
}

func (ir *InventoryRetrievalRepositoryImpl) RetrieveInventory(vault *string, jobQueue *string, initialWaitTime time.Duration, waitInterval time.Duration) (*string, error) {
	if initialWaitTime > 0 {
		time.Sleep(initialWaitTime)
	}

	jobId, err := ir.svc.InitiateInventoryRetrievalJob(vault)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Inventory-retrieval job started with id: %s\n", *jobId)

	notif, err := ir.jobNotificationHandler.GetNotification(jobQueue, waitInterval)
	if err != nil {
		return nil, err
	}

	fmt.Println("Job completion notification received: ", *notif)
	return ir.svc.GetJobOutput(jobId, vault)
}
