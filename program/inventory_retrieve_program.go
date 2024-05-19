package program

import (
	"flag"
	"fmt"
	"s3glacier-go/adapter"
	"s3glacier-go/domain"
	svc "s3glacier-go/svc"
	"time"
)

type InventoryRetrieveProgram struct {
	vault                string
	initialWaitTimeInHrs int
}

func (p *InventoryRetrieveProgram) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to retrieve inventory from")
	fs.IntVar(&p.initialWaitTimeInHrs, "w", 3, "Number of hours to wait before querying job status, default to 3 since S3 jobs are ready in 3-5 hrs")
}

func (p *InventoryRetrieveProgram) Run() error {
	sqsSvc := createSqsClient()
	notif := adapter.NewNotificationHandler(sqsSvc)

	s3g := createGlacierClient()
	csp := adapter.NewCloudServiceProvider(s3g)

	rtrvSvc := svc.NewInventoryRetrieveService(notif, csp)
	initialWaitTime := time.Duration(int64(p.initialWaitTimeInHrs) * int64(time.Hour))

	notificationQueue := domain.NOTIF_QUEUE_NAME
	inv, err := rtrvSvc.RetrieveInventory(&p.vault, &notificationQueue, initialWaitTime, domain.DefaultWaitInterval)
	if err != nil {
		return fmt.Errorf("unable to retrieve inventory: %w", err)
	}

	fmt.Printf("Inventory retrieved:\n%s", *inv)
	return nil
}
