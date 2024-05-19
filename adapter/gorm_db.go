package adapter

import (
	"fmt"
	"s3glacier-go/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type gormDB struct {
	db *gorm.DB
}

func NewDBDAO(connStr string, schemaName string) domain.DBDAO {
	db, err := gorm.Open(
		postgres.Open(connStr),
		&gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix: schemaName + ".",
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

func (dao *gormDB) GetUploadByID(id uint) (*domain.Upload, error) {
	var upload domain.Upload
	if err := dao.db.Find(&upload, id).Error; err != nil {
		return nil, fmt.Errorf("unable to get upload by id %d: %w", id, err)
	}
	return &upload, nil
}

func (dao *gormDB) GetMaxSegNumByUploadID(id uint) (int, error) {
	var maxSegNum int
	if err := dao.db.Raw(
		"SELECT max(segment_num) FROM uploaded_segments WHERE upload_id = ?", id,
	).Scan(&maxSegNum).Error; err != nil {
		return 0, fmt.Errorf("unable to select the max seg num for upload %d: %w", id, err)
	}

	return maxSegNum, nil
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

func (dao *gormDB) UpdateUpload(upload *domain.Upload) error {
	return dao.db.Save(upload).Error
}

func (dao *gormDB) InsertUploadedSegment(seg *domain.UploadedSegment) error {
	return dao.db.Create(seg).Error
}

func (dao *gormDB) GetDownloadByID(id uint) (*domain.Download, error) {
	var dl domain.Download
	if err := dao.db.Find(&dl, id).Error; err != nil {
		return nil, fmt.Errorf("unable to get downloads by id %d: %w", id, err)
	}
	return &dl, nil
}

func (dao *gormDB) InsertDownload(dl *domain.Download) error {
	return dao.db.Create(dl).Error
}

func (dao *gormDB) UpdateDownload(dl *domain.Download) error {
	return dao.db.Save(dl).Error
}

func (dao *gormDB) InsertDownloadedSegment(seg *domain.DownloadedSegment) error {
	return dao.db.Create(seg).Error
}
