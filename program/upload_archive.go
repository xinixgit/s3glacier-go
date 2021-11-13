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
	uploadId uint
}

func (ar *UploadArchive) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&ar.vault, "v", "", "The name of the vault to upload the archive to")
	fs.StringVar(&ar.fpat, "f", "", "The regex of the archive files to be uploaded, you can use `*` to upload all files in a folder, or specify a single file")
	fs.StringVar(&ar.dbuser, "u", "", "The username of the MySQL database")
	fs.StringVar(&ar.dbpwd, "p", "", "The password of the MySQL database")
	fs.StringVar(&ar.dbname, "db", "", "The name of the database created")
	fs.StringVar(&ar.dbip, "ip", "localhost:3306", "The IP address and port number of the database, default to `localhost:3306`")
	fs.UintVar(&ar.uploadId, "uploadId", 0, "The id of the upload (from the `uploads` table) to resume, if some of its parts had failed to be uploaded previously")
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
		resumedUpload := dbdao.GetUploadByID(ar.uploadId)
		maxSegNum := dbdao.GetMaxSegNumByUploadID(ar.uploadId)
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
