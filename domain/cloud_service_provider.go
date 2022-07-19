package domain

type CloudServiceProvider interface {
	DeleteArchive(accountID *string, archiveID *string, vaultName *string) error
}
