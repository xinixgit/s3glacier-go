package adapter

import (
	"s3glacier-go/domain"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type GormDB struct {
	db *gorm.DB
}

func NewDBDAO(connStr string) domain.DBDAO {
	db, err := gorm.Open(mysql.Open(connStr), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &GormDB{
		db: db,
	}
}

func (dao *GormDB) GetUploadByID(id uint) *domain.Upload {
	var upload domain.Upload
	dao.db.Find(&upload, id)
	return &upload
}

func (dao *GormDB) GetMaxSegNumByUploadID(id uint) int {
	var maxSegNum int
	dao.db.Raw("SELECT max(segment_num) FROM uploaded_segments WHERE upload_id = ?", id).Scan(&maxSegNum)
	return maxSegNum
}

func (dao *GormDB) GetExpiredUpload(vault *string) ([]domain.Upload, error) {
	var uploads []domain.Upload
	txn := dao.db.Where(
		"CAST(created_at AS DATE) < DATE_SUB(NOW(), INTERVAL 4 MONTH) AND vault_name = ? AND archive_id IS NOT NULL AND archive_id != '' AND status != ?",
		*vault,
		domain.DELETED,
	).Find(&uploads)
	return uploads, txn.Error
}

func (dao *GormDB) InsertUpload(upload *domain.Upload) error {
	return dao.db.Create(upload).Error
}

func (dao *GormDB) UpdateUpload(upload *domain.Upload) {
	dao.db.Save(upload)
}

func (dao *GormDB) InsertUploadedSegment(seg *domain.UploadedSegment) error {
	return dao.db.Create(seg).Error
}

func (dao *GormDB) GetDownloadByID(id uint) *domain.Download {
	var dl domain.Download
	dao.db.Find(&dl, id)
	return &dl
}

func (dao *GormDB) InsertDownload(dl *domain.Download) error {
	return dao.db.Create(dl).Error
}

func (dao *GormDB) UpdateDownload(dl *domain.Download) {
	dao.db.Save(dl)
}

func (dao *GormDB) InsertDownloadedSegment(seg *domain.DownloadedSegment) error {
	return dao.db.Create(seg).Error
}
