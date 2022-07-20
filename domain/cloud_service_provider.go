package domain

import (
	"io"
	"time"
)

type JobDescription struct {
	Completed          *bool
	ArchiveId          *string
	JobId              *string
	ArchiveSizeInBytes *int64
}

type JobOutput struct {
	Body     io.ReadCloser
	Checksum *string
}

type CloudServiceProvider interface {
	DeleteArchive(archiveID *string, vaultName *string) error
	DescribeJob(jobId *string, vaultName *string) (*JobDescription, error)
	GetJobOutput(jobId *string, vaultName *string) (*JobOutput, error)
	GetJobOutputByRange(jobId *string, bytesRange *string, vaultName *string) (*JobOutput, error)
	InitiateArchiveRetrievalJob(archiveID *string, vault *string) (*string, error)
	InitiateInventoryRetrievalJob(vault *string) (*string, error)
	// Actively check the status of a job, and execute the function when the job is completed
	OnJobComplete(jobID *string, archiveId *string, vault *string, waitInterval time.Duration, onComplete func(int)) error
}
