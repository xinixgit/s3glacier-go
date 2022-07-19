package adapter

import (
	"s3glacier-go/domain"

	"github.com/aws/aws-sdk-go/service/glacier"
)

type CloudServiceProviderImpl struct {
	s3g *glacier.Glacier
}

func NewCloudServiceProvider(s3g *glacier.Glacier) domain.CloudServiceProvider {
	return &CloudServiceProviderImpl{
		s3g: s3g,
	}
}

func (svc *CloudServiceProviderImpl) DeleteArchive(accountID *string, archiveID *string, vaultName *string) error {
	svcInput := &glacier.DeleteArchiveInput{
		AccountId: accountID,
		ArchiveId: archiveID,
		VaultName: vaultName,
	}

	_, err := svc.s3g.DeleteArchive(svcInput)
	return err
}
