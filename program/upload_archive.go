package program

import (
	"flag"
	"fmt"
	"path/filepath"
	"xddd/s3glacier/db"
	"xddd/s3glacier/upload"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
)

type UploadArchive struct {
	vault    string
	fpat     string
	dbuser   string
	dbpwd    string
	dbname   string
	dbip     string
	uploadId int
}

func (ar *UploadArchive) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&ar.vault, "v", "", "the name of the vault to upload data into")
	fs.StringVar(&ar.fpat, "f", "", "the regex of files to be uploaded")
	fs.StringVar(&ar.dbuser, "u", "", "the username of the database")
	fs.StringVar(&ar.dbpwd, "p", "", "the password of the database")
	fs.StringVar(&ar.dbname, "db", "", "the name of the database")
	fs.StringVar(&ar.dbip, "ip", "localhost:3306", "the ip address and port of the database")
	fs.IntVar(&ar.uploadId, "uploadId", 0, "the id of the upload to resume uploading, if partially failed previously")
}

func (ar *UploadArchive) Run() {
	files, err := filepath.Glob(ar.fpat)
	if err != nil {
		panic(err)
	}

	if len(files) > 1 && ar.uploadId > 0 {
		panic("Seg number only works when uploading a single file.")
	}

	s3glacier := createGlacierClient()
	dbdao := db.NewDBDAO(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", ar.dbuser, ar.dbpwd, ar.dbip, ar.dbname))
	uploader := upload.S3GlacierUploader{Vault: &ar.vault, S3glacier: s3glacier, DBDAO: dbdao}

	if ar.uploadId > 0 {
		resumedUpload := dbdao.GetUploadByID(uint(ar.uploadId))
		maxSegNum := dbdao.GetMaxSegNumByUploadID(uint(ar.uploadId))
		uploader.ResumedUpload = &upload.ResumedUpload{Upload: resumedUpload, MaxSegNum: maxSegNum}
	}

	for _, f := range files {
		uploader.Upload(f)
	}
}

func createGlacierClient() *glacier.Glacier {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})

	if err != nil {
		panic(err)
	}

	return glacier.New(sess)
}
