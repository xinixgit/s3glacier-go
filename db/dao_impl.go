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

func (dao *DBDAOImpl) GetMaxSegNumByUploadID(id uint) int {
	var maxSegNum int
	dao.db.Raw("SELECT max(upload_id) FROM uploaded_segments WHERE upload_id = ?", id).Scan(&maxSegNum)
	return maxSegNum
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
