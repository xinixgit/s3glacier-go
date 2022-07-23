package program

import (
	"flag"
	"fmt"
	"s3glacier-go/adapter"
	"s3glacier-go/app"
	"s3glacier-go/domain"
	"time"
)

type InventoryRetrieval struct {
	vault                string
	initialWaitTimeInHrs int
}

func (p *InventoryRetrieval) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to retrieve inventory from")
	fs.IntVar(&p.initialWaitTimeInHrs, "w", 3, "Number of hours to wait before querying job status, default to 3 since S3 jobs are ready in 3-5 hrs")
}

func (p *InventoryRetrieval) Run() {
	sqsSvc := CreateSqsClient()
	h := adapter.NewJobNotificationHandler(sqsSvc)

	s3g := CreateGlacierClient()
	svc := adapter.NewCloudServiceProvider(s3g)

	repo := app.NewInventoryRetrievalRepository(h, svc)
	initialWaitTime := time.Duration(int64(p.initialWaitTimeInHrs) * int64(time.Hour))

	notificationQueue := domain.NOTIF_QUEUE_NAME
	inv, err := repo.RetrieveInventory(&p.vault, &notificationQueue, initialWaitTime, domain.DefaultWaitInterval)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inventory retrieved:\n%s", *inv)
}
