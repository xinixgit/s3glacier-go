package program

import (
	"flag"
	"fmt"
	"s3glacier-go/inventoryretrieval"
	"s3glacier-go/util"
	"time"
)

const TEN_MINUTE = 10 * time.Minute

const NOTIF_QUEUE_NAME = string("glacier-job-notif-queue")

type InventoryRetrieval struct {
	vault                string
	initialWaitTimeInHrs int
}

func (p *InventoryRetrieval) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to retrieve inventory from")
	fs.IntVar(&p.initialWaitTimeInHrs, "w", 3, "Number of hours to wait before querying job status, default to 3 since S3 jobs are ready in 3-5 hrs")
}

func (p *InventoryRetrieval) Run() {
	s3glacier := CreateGlacierClient()

	q := NOTIF_QUEUE_NAME
	handler := &util.JobNotificationHandler{
		QueueName:     &q,
		Svc:           CreateSqsClient(),
		SleepInterval: TEN_MINUTE,
	}

	ir := inventoryretrieval.S3GlacierInventoryRetrieval{
		Vault:                  &p.vault,
		InitialWaitTime:        time.Duration(int64(p.initialWaitTimeInHrs) * int64(time.Hour)),
		JobNotificationHandler: handler,
		S3glacier:              s3glacier,
	}

	inv, err := ir.RetrieveInventory()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inventory retrieved:\n%s", *inv)
}
