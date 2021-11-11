package program

import (
	"fmt"
	"flag"
	"path/filepath"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/glacier"
	"xddd/s3glacier/upload"
	"xddd/s3glacier/db"
)

type UploadArchive struct {}

func (p *UploadArchive) Run() {
	vault 	:= flag.String("v", "", "the name of the vault to upload data into")
	fpat 		:= flag.String("f", "", "the regex of files to be uploaded")
	dbuser 	:= flag.String("u", "", "the username of the database")
	dbpwd 	:= flag.String("p", "", "the password of the database")
	dbname 	:= flag.String("db", "", "the name of the database")
	dbip 		:= flag.String("ip", "localhost:3306", "the ip address and port of the database")

	flag.Parse()

	flag.VisitAll(func (f *flag.Flag) {
		if f.Name != "ip" && f.Value.String() == "" {
			fmt.Printf("Usage: s3glacier upload-archive\n")
			flag.PrintDefaults()
			panic("End execution now.")
		}
	})

	files, err := filepath.Glob(*fpat)
  if err != nil {
		panic(err)
	}

	s3glacier := createGlacierClient()
	dbdao := db.NewDBDAO(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", *dbuser, *dbpwd, *dbip, *dbname))
	uploader := upload.S3GlacierUploader{Vault: vault, S3glacier: s3glacier, DBDAO: dbdao}

	for _, f := range files {
		uploader.Upload(f)
	}
}

func createGlacierClient() *glacier.Glacier {
	sess, err := session.NewSession(&aws.Config{
    Region: 			aws.String("us-west-2"),
  })

  if err != nil {
  	panic(err)
  }

	return glacier.New(sess)
}