package domain

type Status int

const (
	COMPLETED Status = iota
	STARTED
	FAILED
	DELETED
)

type Upload struct {
	ID        uint `gorm:"column:id";primaryKey`
	VaultName string
	Filename  string
	CreatedAt string
	DeletedAt string
	Location  string
	SessionId string
	Checksum  string
	ArchiveId string
	Status    Status
}

type UploadedSegment struct {
	ID         uint `gorm:"column:id";primaryKey`
	SegmentNum int
	UploadId   uint
	Checksum   string
	CreatedAt  string
}

type Download struct {
	ID        uint `gorm:"column:id";primaryKey`
	VaultName string
	JobId     string
	ArchiveId string
	CreatedAt string
	UpdatedAt string
	Status    Status
}

type DownloadedSegment struct {
	ID         uint `gorm:"column:id";primaryKey`
	BytesRange string
	DownloadId uint
	CreatedAt  string
}

type DBDAO interface {
	GetUploadByID(id uint) *Upload
	GetMaxSegNumByUploadID(id uint) int
	GetExpiredUpload(vault *string) ([]Upload, error)
	InsertUpload(upload *Upload) error
	UpdateUpload(upload *Upload)
	InsertUploadedSegment(seg *UploadedSegment) error
	GetDownloadByID(id uint) *Download
	InsertDownload(dl *Download) error
	UpdateDownload(dl *Download)
	InsertDownloadedSegment(seg *DownloadedSegment) error
}
