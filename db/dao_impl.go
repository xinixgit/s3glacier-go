package db

import (
	"gorm.io/gorm"
)

type UploadDAOImpl struct {
	db *gorm.DB
}

func (dao *UploadDAOImpl) GetUploadByID(id uint) *Upload {
	var upload Upload
	dao.db.Find(&upload, id)
	return &upload
}

func (dao *UploadDAOImpl) GetMaxSegNumByUploadID(id uint) int {
	var maxSegNum int
	dao.db.Raw("SELECT max(segment_num) FROM uploaded_segments WHERE upload_id = ?", id).Scan(&maxSegNum)
	return maxSegNum
}

func (dao *UploadDAOImpl) GetExpiredUpload(vault *string) ([]Upload, error) {
	var uploads []Upload
	txn := dao.db.Where(
		"CAST(created_at AS DATE) < DATE_SUB(NOW(), INTERVAL 4 MONTH) AND vault_name = ? AND archive_id IS NOT NULL AND archive_id != '' AND status != ?", *vault, DELETED).Find(&uploads)
	return uploads, txn.Error
}

func (dao *UploadDAOImpl) InsertUpload(upload *Upload) error {
	return dao.db.Create(upload).Error
}

func (dao *UploadDAOImpl) UpdateUpload(upload *Upload) {
	dao.db.Save(upload)
}

func (dao *UploadDAOImpl) InsertUploadedSegment(seg *UploadedSegment) error {
	return dao.db.Create(seg).Error
}

type DownloadDAOImpl struct {
	db *gorm.DB
}

func (dao *DownloadDAOImpl) GetDownloadByID(id uint) *Download {
	var dl Download
	dao.db.Find(&dl, id)
	return &dl
}

func (dao *DownloadDAOImpl) InsertDownload(dl *Download) error {
	return dao.db.Create(dl).Error
}

func (dao *DownloadDAOImpl) UpdateDownload(dl *Download) {
	dao.db.Save(dl)
}

func (dao *DownloadDAOImpl) InsertDownloadedSegment(seg *DownloadedSegment) error {
	return dao.db.Create(seg).Error
}
