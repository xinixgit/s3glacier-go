package adapter

import (
	"io"
	"s3glacier-go/domain"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

type AWSGlacierCloudServiceProvider struct {
	s3g *glacier.Glacier
}

func NewCloudServiceProvider(s3g *glacier.Glacier) domain.CloudServiceProvider {
	return &AWSGlacierCloudServiceProvider{
		s3g: s3g,
	}
}

func (svc *AWSGlacierCloudServiceProvider) InitiateInventoryRetrievalJob(vault *string) (*string, error) {
	input := &glacier.InitiateJobInput{
		AccountId: aws.String("-"),
		JobParameters: &glacier.JobParameters{
			Type:   aws.String("inventory-retrieval"),
			Format: aws.String("JSON"),
		},
		VaultName: vault,
	}

	res, err := svc.s3g.InitiateJob(input)
	if err != nil {
		return nil, err
	}
	return res.JobId, nil
}

func (svc *AWSGlacierCloudServiceProvider) DeleteArchive(archiveID *string, vaultName *string) error {
	svcInput := &glacier.DeleteArchiveInput{
		AccountId: aws.String("-"),
		ArchiveId: archiveID,
		VaultName: vaultName,
	}

	_, err := svc.s3g.DeleteArchive(svcInput)
	return err
}

func (svc *AWSGlacierCloudServiceProvider) GetJobOutput(jobId *string, vault *string) (*string, error) {
	input := &glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		JobId:     jobId,
		VaultName: vault,
	}

	res, err := svc.s3g.GetJobOutput(input)
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
