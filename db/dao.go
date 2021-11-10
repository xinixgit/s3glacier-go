package db

import (
  "gorm.io/gorm"
  "gorm.io/driver/mysql"
)

type DBDAO interface {
	GetUploadByID(id uint) *Upload
	InsertUpload(upload *Upload) error
	UpdateUpload(upload *Upload)
	InsertUploadedSegment(seg *UploadedSegment) error
}

func NewDBDAO(connStr string) DBDAO {
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &DBDAOImpl{db}
}
