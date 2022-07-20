package adapter

import (
	"errors"
	"fmt"
	"s3glacier-go/domain"
	"time"

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

func (svc *AWSGlacierCloudServiceProvider) InitiateArchiveRetrievalJob(archiveID *string, vault *string) (*string, error) {
	input := &glacier.InitiateJobInput{
		AccountId: aws.String("-"),
		JobParameters: &glacier.JobParameters{
			ArchiveId: archiveID,
			Type:      aws.String("archive-retrieval"),
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

func (svc *AWSGlacierCloudServiceProvider) GetJobOutput(jobId *string, vault *string) (*domain.JobOutput, error) {
	input := &glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		JobId:     jobId,
		VaultName: vault,
	}

	return svc.getJobOutput(input)
}

func (svc *AWSGlacierCloudServiceProvider) GetJobOutputByRange(jobId *string, bytesRange *string, vault *string) (*domain.JobOutput, error) {
	input := &glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		JobId:     jobId,
		Range:     bytesRange,
		VaultName: vault,
	}

	return svc.getJobOutput(input)
}

func (svc *AWSGlacierCloudServiceProvider) DescribeJob(jobId *string, vaultName *string) (*domain.JobDescription, error) {
	input := &glacier.DescribeJobInput{
		AccountId: aws.String("-"),
		JobId:     jobId,
		VaultName: vaultName,
	}
	res, err := svc.s3g.DescribeJob(input)
	if err != nil {
		return nil, err
	}

	return &domain.JobDescription{
		Completed:          res.Completed,
		ArchiveId:          res.ArchiveId,
		JobId:              res.JobId,
		ArchiveSizeInBytes: res.ArchiveSizeInBytes,
	}, nil
}

func (svc *AWSGlacierCloudServiceProvider) OnJobComplete(
	jobID *string,
	archiveId *string,
	vault *string,
	waitInterval time.Duration,
	onComplete func(int),
) error {
	for {
		fmt.Printf("Wait %ds before next job status poll.\n", int(waitInterval.Seconds()))
		time.Sleep(waitInterval)

		res, err := svc.DescribeJob(jobID, vault)

		if err != nil {
			fmt.Println("Failed to pull job status from s3, ", err)
			return err
		}
		if !*res.Completed {
			fmt.Println("Job is not ready.")
			continue
		}
		if arId := res.ArchiveId; *arId != *archiveId {
			msg := fmt.Sprintf("Archive id not matching! Expected: %s, received: %s", *archiveId, *arId)
			return errors.New(msg)
		}
		if jId := res.JobId; *jId != *jobID {
			msg := fmt.Sprintf("Job id not matching! Expected: %s, received: %s", *jobID, *jId)
			return errors.New(msg)
		}

		sizeInBytes := res.ArchiveSizeInBytes
		fmt.Printf("Now start downloading the file of %d bytes.\n", *sizeInBytes)
		onComplete(int(*sizeInBytes))

		break
	}

	return nil
}

func (svc *AWSGlacierCloudServiceProvider) getJobOutput(input *glacier.GetJobOutputInput) (*domain.JobOutput, error) {
	res, err := svc.s3g.GetJobOutput(input)
	if err != nil {
		return nil, err
	}

	return &domain.JobOutput{
		Body:     res.Body,
		Checksum: res.Checksum,
	}, nil
}
