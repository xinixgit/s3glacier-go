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
	vault  string
	fpat   string
	dbuser string
	dbpwd  string
	dbname string
	dbip   string
	partno int
}

func (ar *UploadArchive) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&ar.vault, "v", "", "the name of the vault to upload data into")
	fs.StringVar(&ar.fpat, "f", "", "the regex of files to be uploaded")
	fs.StringVar(&ar.dbuser, "u", "", "the username of the database")
	fs.StringVar(&ar.dbpwd, "p", "", "the password of the database")
	fs.StringVar(&ar.dbname, "db", "", "the name of the database")
	fs.StringVar(&ar.dbip, "ip", "localhost:3306", "the ip address and port of the database")
	fs.IntVar(&ar.partno, "partno", 0, "the part number of the archive to resume uploading from, if not uploading from the start of the file")
}

func (ar *UploadArchive) Run() {
	files, err := filepath.Glob(ar.fpat)
	if err != nil {
		panic(err)
	}

	if len(files) > 1 && ar.partno > 1 {
		panic("Partno only works when uploading a single file.")
	}

	s3glacier := createGlacierClient()
	dbdao := db.NewDBDAO(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", ar.dbuser, ar.dbpwd, ar.dbip, ar.dbname))
	uploader := upload.S3GlacierUploader{Vault: &ar.vault, S3glacier: s3glacier, DBDAO: dbdao, Partno: ar.partno}

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
