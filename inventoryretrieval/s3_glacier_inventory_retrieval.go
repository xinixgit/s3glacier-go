package inventoryretrieval

import (
	"fmt"
	"io"
	"s3glacier-go/util"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

const TEN_MINUTE = 10 * time.Minute

type S3GlacierInventoryRetrieval struct {
	Vault                  *string
	InitialWaitTime        time.Duration
	JobNotificationHandler *util.JobNotificationHandler
	S3glacier              *glacier.Glacier
}

func (ir S3GlacierInventoryRetrieval) RetrieveInventory() {
	jobId := ir.initiateInventoryRetrievalJob()
	fmt.Printf("Inventory-retrieval job started with id: %s\n", *jobId)
	notif, err := ir.JobNotificationHandler.GetNotification()
	if err != nil {
		panic(err)
	}
	fmt.Println("Job completion notification received: ", *notif)

	inv, err := ir.retrievePreparedInventory(jobId)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inventory retrieved:\n%s", *inv)
}

func (ir S3GlacierInventoryRetrieval) initiateInventoryRetrievalJob() *string {
	input := &glacier.InitiateJobInput{
		AccountId: aws.String("-"),
		JobParameters: &glacier.JobParameters{
			Type:   aws.String("inventory-retrieval"),
			Format: aws.String("JSON"),
		},
		VaultName: ir.Vault,
	}

	res, err := ir.S3glacier.InitiateJob(input)
	if err != nil {
		panic(err)
	}
	return res.JobId
}

func (ir S3GlacierInventoryRetrieval) retrievePreparedInventory(jobId *string) (*string, error) {
	input := &glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		JobId:     jobId,
		VaultName: ir.Vault,
	}

	res, err := ir.S3glacier.GetJobOutput(input)
	if err != nil {
		return nil, err
	}

	buf := new(strings.Builder)
	if _, err2 := io.Copy(buf, res.Body); err2 != nil {
		return nil, err
	}

	inv := buf.String()
	return &inv, nil
}
