package domain

type CloudServiceProvider interface {
	DeleteArchive(archiveID *string, vaultName *string) error
	InitiateInventoryRetrievalJob(vault *string) (*string, error)
	GetJobOutput(jobId *string, vaultName *string) (*string, error)
}
