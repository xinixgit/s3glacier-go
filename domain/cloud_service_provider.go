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

type UploadJobOutput struct {
}

type CompleteMultipartUploadOutput struct {
	Location  *string
	Checksum  *string
	ArchiveID *string
}

type CloudServiceProvider interface {
	AbortMultipartUpload(sessionID *string, vault *string) error
	CompleteMultipartUploadInput(archiveSize int64, checksum *string, sessionID *string, vault *string) (*CompleteMultipartUploadOutput, error)
	DeleteArchive(archiveID *string, vaultName *string) error
	DescribeJob(jobId *string, vaultName *string) (*JobDescription, error)
	GetJobOutput(jobId *string, vaultName *string) (*JobOutput, error)
	GetJobOutputByRange(jobId *string, bytesRange *string, vaultName *string) (*JobOutput, error)
	InitiateArchiveRetrievalJob(archiveID *string, vault *string) (*string, error)
	InitiateInventoryRetrievalJob(vault *string) (*string, error)
	InitiateMultipartUpload(chunkSize int, vault *string) (*string, error)
	// Actively check the status of a job, and execute the function when the job is completed
	OnJobComplete(jobID *string, archiveId *string, vault *string, waitInterval time.Duration, onComplete func(int)) error
	UploadMultipartPart(segment []byte, checksum *string, byteRange *string, sessionID *string, vault *string) (*UploadJobOutput, error)
}
