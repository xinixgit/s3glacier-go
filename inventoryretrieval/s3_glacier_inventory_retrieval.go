package inventoryretrieval

import (
	"fmt"
	"s3glacier-go/util"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

const TEN_MINUTE = 10 * time.Minute

type S3GlacierInventoryRetrieval struct {
	Vault           *string
	InitialWaitTime time.Duration
	S3glacier       *glacier.Glacier
}

func (ir S3GlacierInventoryRetrieval) RetrieveInventory() {
	jobId := ir.initiateInventoryRetrievalJob()
	util.ListenForJobOutput(ir.Vault, jobId, ir.onJobComplete, ir.InitialWaitTime, TEN_MINUTE, ir.S3glacier)
}

func (ir S3GlacierInventoryRetrieval) initiateInventoryRetrievalJob() *string {
	input := &glacier.InitiateJobInput{
		AccountId: aws.String("-"),
		JobParameters: &glacier.JobParameters{
			Type: aws.String("inventory-retrieval"),
		},
		VaultName: ir.Vault,
	}

	res, err := ir.S3glacier.InitiateJob(input)
	if err != nil {
		panic(err)
	}
	return res.JobId
}

func (ir S3GlacierInventoryRetrieval) onJobComplete(desc *glacier.JobDescription) {
	fmt.Println(desc)
}
