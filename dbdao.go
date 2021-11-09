package main

import (
	"gorm.io/driver/mysql"
  "gorm.io/gorm"
)

type DBDAO interface {
	GetUploadByID(id uint) *Upload
	InsertUpload(upload *Upload) error
	UpdateUpload(upload *Upload)
	InsertUploadedSegment(seg *UploadedSegment) error
}

type DBDAOImpl struct {
	db *gorm.DB
}

func NewDBDAO(connStr string) DBDAO {
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &DBDAOImpl{db}
}

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

func (dao *DBDAOImpl) GetUploadByID(id uint) *Upload {
	var upload Upload
	dao.db.Find(&upload, id)
	return &upload
}

func (dao *DBDAOImpl) InsertUpload(upload *Upload) error {
	res := dao.db.Create(upload)
	return res.Error
}

func (dao *DBDAOImpl) UpdateUpload(upload *Upload) {
	dao.db.Save(upload)
}

func (dao *DBDAOImpl) InsertUploadedSegment(seg *UploadedSegment) error {
	res := dao.db.Create(seg)
	return res.Error
}


































