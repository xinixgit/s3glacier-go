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
	GetDownloadByID(id uint) *Download
	GetExpiredUpload(vault *string) ([]Upload, error)
	GetMaxSegNumByUploadID(id uint) int
	GetUploadByID(id uint) *Upload
	InsertDownload(dl *Download) error
	InsertDownloadedSegment(seg *DownloadedSegment) error
	InsertUpload(upload *Upload) error
	InsertUploadedSegment(seg *UploadedSegment) error
	UpdateDownload(dl *Download)
	UpdateUpload(upload *Upload)
}
