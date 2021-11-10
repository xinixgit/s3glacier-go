package db

type Status int

const (
	COMPLETED Status = iota
	STARTED
	FAILED
)

type Upload struct {
	ID 					uint 		`gorm:"column:id";primaryKey`
	VaultName		string
	Filename		string
	CreatedAt 	string
	Location 		string
	SessionId 	string
	Checksum 		string
	ArchiveId 	string
	Status 			Status
}

type UploadedSegment struct {
	ID 					uint 		`gorm:"column:id";primaryKey`
	SegmentNum 	int
	UploadId 		uint
	Checksum  	string
	CreatedAt 	string
}
