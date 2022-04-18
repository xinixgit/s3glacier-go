package db

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type UploadDAO interface {
	GetUploadByID(id uint) *Upload
	GetMaxSegNumByUploadID(id uint) int
	GetExpiredUpload(vault *string) ([]Upload, error)
	InsertUpload(upload *Upload) error
	UpdateUpload(upload *Upload)
	InsertUploadedSegment(seg *UploadedSegment) error
}

type DownloadDAO interface {
	GetDownloadByID(id uint) *Download
	InsertDownload(dl *Download) error
	UpdateDownload(dl *Download)
	InsertDownloadedSegment(seg *DownloadedSegment) error
}

func NewUploadDAO(connStr string) UploadDAO {
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &UploadDAOImpl{db}
}

func NewDownloadDAO(connStr string) DownloadDAO {
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &DownloadDAOImpl{db}
}
