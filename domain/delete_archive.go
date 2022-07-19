package domain

type Archive struct {
	ArchiveId          string
	ArchiveDescription string
	CreationDate       string
	Size               int64
	SHA256TreeHash     string
}

type Inventory struct {
	VaultARN      string
	InventoryDate string
	ArchiveList   []Archive
}
