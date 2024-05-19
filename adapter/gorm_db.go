package adapter

import (
	"s3glacier-go/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type gormDB struct {
	db *gorm.DB
}

func NewDBDAO(connStr string) domain.DBDAO {
	db, err := gorm.Open(
		postgres.Open(connStr),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: "s3g.",
			},
		},
	)
	if err != nil {
		panic(err)
	}
	return &gormDB{
		db: db,
	}
}

func (dao *gormDB) GetUploadByID(id uint) *domain.Upload {
	var upload domain.Upload
	dao.db.Find(&upload, id)
	return &upload
}

func (dao *gormDB) GetMaxSegNumByUploadID(id uint) int {
	var maxSegNum int
	dao.db.Raw("SELECT max(segment_num) FROM uploaded_segments WHERE upload_id = ?", id).Scan(&maxSegNum)
	return maxSegNum
}

func (dao *gormDB) GetExpiredUpload(vault *string) ([]domain.Upload, error) {
	var uploads []domain.Upload
	txn := dao.db.Where(
		"CAST(created_at AS DATE) < DATE_SUB(NOW(), INTERVAL 4 MONTH) AND vault_name = ? AND archive_id IS NOT NULL AND archive_id != '' AND status = ?",
		*vault,
		domain.COMPLETED,
	).Find(&uploads)
	return uploads, txn.Error
}

func (dao *gormDB) InsertUpload(upload *domain.Upload) error {
	return dao.db.Create(upload).Error
}

func (dao *gormDB) UpdateUpload(upload *domain.Upload) {
	dao.db.Save(upload)
}

func (dao *gormDB) InsertUploadedSegment(seg *domain.UploadedSegment) error {
	return dao.db.Create(seg).Error
}

func (dao *gormDB) GetDownloadByID(id uint) *domain.Download {
	var dl domain.Download
	dao.db.Find(&dl, id)
	return &dl
}

func (dao *gormDB) InsertDownload(dl *domain.Download) error {
	return dao.db.Create(dl).Error
}

func (dao *gormDB) UpdateDownload(dl *domain.Download) {
	dao.db.Save(dl)
}

func (dao *gormDB) InsertDownloadedSegment(seg *domain.DownloadedSegment) error {
	return dao.db.Create(seg).Error
}
