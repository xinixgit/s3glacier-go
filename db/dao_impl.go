package db

import (
	"gorm.io/gorm"
)

type DBDAOImpl struct {
	db *gorm.DB
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
