package adapter

import (
	"bytes"
	"s3glacier-go/domain"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

type AWSS3Glacier struct {
	s3g *glacier.Glacier
}

func NewCloudServiceProvider(s3g *glacier.Glacier) domain.CloudServiceProvider {
	return &AWSS3Glacier{
		s3g: s3g,
	}
}

func (svc *AWSS3Glacier) InitiateInventoryRetrievalJob(vault *string) (*string, error) {
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

func (svc *AWSS3Glacier) InitiateArchiveRetrievalJob(archiveID *string, vault *string) (*string, error) {
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

func (svc *AWSS3Glacier) DeleteArchive(archiveID *string, vaultName *string) error {
	svcInput := &glacier.DeleteArchiveInput{
		AccountId: aws.String("-"),
		ArchiveId: archiveID,
		VaultName: vaultName,
	}

	_, err := svc.s3g.DeleteArchive(svcInput)
	return err
}

func (svc *AWSS3Glacier) GetJobOutput(jobId *string, vault *string) (*domain.JobOutput, error) {
	input := &glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		JobId:     jobId,
		VaultName: vault,
	}

	return svc.getJobOutput(input)
}

func (svc *AWSS3Glacier) GetJobOutputByRange(jobId *string, bytesRange *string, vault *string) (*domain.JobOutput, error) {
	input := &glacier.GetJobOutputInput{
		AccountId: aws.String("-"),
		JobId:     jobId,
		Range:     bytesRange,
		VaultName: vault,
	}

	return svc.getJobOutput(input)
}

func (svc *AWSS3Glacier) DescribeJob(jobId *string, vaultName *string) (*domain.JobDescription, error) {
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
		SNSTopic:           res.SNSTopic,
		StatusCode:         res.StatusCode,
		CreationDate:       res.CreationDate,
	}, nil
}

func (svc *AWSS3Glacier) getJobOutput(input *glacier.GetJobOutputInput) (*domain.JobOutput, error) {
	res, err := svc.s3g.GetJobOutput(input)
	if err != nil {
		return nil, err
	}

	return &domain.JobOutput{
		Body:     res.Body,
		Checksum: res.Checksum,
	}, nil
}

func (svc *AWSS3Glacier) InitiateMultipartUpload(chunkSize int, vault *string) (*string, error) {
	input := &glacier.InitiateMultipartUploadInput{
		AccountId: aws.String("-"),
		PartSize:  aws.String(strconv.Itoa(chunkSize)),
		VaultName: vault,
	}

	out, err := svc.s3g.InitiateMultipartUpload(input)
	if err != nil {
		return nil, err
	}
	return out.UploadId, nil
}

func (svc *AWSS3Glacier) UploadMultipartPart(
	segment []byte,
	checksum *string,
	byteRange *string,
	sessionID *string,
	vault *string,
) (*domain.UploadJobOutput, error) {
	input := &glacier.UploadMultipartPartInput{
		AccountId: aws.String("-"),
		Body:      bytes.NewReader(segment),
		Checksum:  checksum,
		Range:     byteRange,
		UploadId:  sessionID,
		VaultName: vault,
	}

	// upload a single segment in a multipart upload session
	if _, err := svc.s3g.UploadMultipartPart(input); err != nil {
		return nil, err
	}

	return &domain.UploadJobOutput{}, nil
}

func (svc *AWSS3Glacier) AbortMultipartUpload(sessionID *string, vault *string) error {
	input := &glacier.AbortMultipartUploadInput{
		AccountId: aws.String("-"),
		UploadId:  sessionID,
		VaultName: vault,
	}

	_, err := svc.s3g.AbortMultipartUpload(input)
	return err
}

func (svc *AWSS3Glacier) CompleteMultipartUploadInput(
	archiveSize int64,
	checksum *string,
	sessionID *string,
	vault *string,
) (*domain.CompleteMultipartUploadOutput, error) {
	input := &glacier.CompleteMultipartUploadInput{
		AccountId:   aws.String("-"),
		ArchiveSize: aws.String(strconv.FormatInt(archiveSize, 10)),
		Checksum:    checksum,
		UploadId:    sessionID,
		VaultName:   vault,
	}
	res, err := svc.s3g.CompleteMultipartUpload(input)
	if err != nil {
		return nil, err
	}

	return &domain.CompleteMultipartUploadOutput{
		Location:  res.Location,
		Checksum:  res.Checksum,
		ArchiveID: res.ArchiveId,
	}, nil
}
