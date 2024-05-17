package domain

import (
	"bytes"
	"fmt"
	"io"
)

type JobDescription struct {
	Completed          *bool
	ArchiveId          *string
	JobId              *string
	ArchiveSizeInBytes *int64
	SNSTopic           *string
	StatusCode         *string
	CreationDate       *string
}

func (jd JobDescription) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("Completed: %t", *jd.Completed))
	if jd.ArchiveId != nil {
		buf.WriteString(fmt.Sprintf("\nArchiveId: %s", *jd.ArchiveId))
	}
	if jd.JobId != nil {
		buf.WriteString(fmt.Sprintf("\nJobId: %s", *jd.JobId))
	}
	if jd.ArchiveSizeInBytes != nil {
		buf.WriteString(fmt.Sprintf("\nArchiveSizeInBytes: %d", *jd.ArchiveSizeInBytes))
	}
	if jd.SNSTopic != nil {
		buf.WriteString(fmt.Sprintf("\nSNSTopic: %s", *jd.SNSTopic))
	}
	if jd.StatusCode != nil {
		buf.WriteString(fmt.Sprintf("\nStatusCode: %s", *jd.StatusCode))
	}
	if jd.CreationDate != nil {
		buf.WriteString(fmt.Sprintf("\nCreationDate: %s", *jd.CreationDate))
	}
	return buf.String()
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
	UploadMultipartPart(segment []byte, checksum *string, byteRange *string, sessionID *string, vault *string) (*UploadJobOutput, error)
}
