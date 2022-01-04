package program

import (
	"flag"
	"s3glacier-go/inventoryretrieval"
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
	s3glacier := CreateGlacierClient()
	ir := inventoryretrieval.S3GlacierInventoryRetrieval{
		Vault:           &p.vault,
		InitialWaitTime: time.Duration(int64(p.initialWaitTimeInHrs) * int64(time.Hour)),
		S3glacier:       s3glacier,
	}

	ir.RetrieveInventory()
}
