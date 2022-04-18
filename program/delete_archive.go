package program

import (
	"flag"
	"fmt"
	"s3glacier-go/db"
	"s3glacier-go/delete"
)

type DeleteArchive struct {
	vault  string
	dbuser string
	dbpwd  string
	dbname string
	dbip   string
}

func (p *DeleteArchive) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to download the archive from")
	fs.StringVar(&p.dbuser, "u", "", "The username of the MySQL database")
	fs.StringVar(&p.dbpwd, "p", "", "The password of the MySQL database")
	fs.StringVar(&p.dbname, "db", "", "The name of the database created")
	fs.StringVar(&p.dbip, "ip", "localhost:3306", "The IP address and port number of the database, default to `localhost:3306`")
}

func (p *DeleteArchive) Run() {
	s3glacier := CreateGlacierClient()
	dbdao := db.NewUploadDAO(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4", p.dbuser, p.dbpwd, p.dbip, p.dbname))

	h := delete.ArchiveDeleteHandler{
		Vault:     &p.vault,
		UlDAO:     dbdao,
		S3glacier: s3glacier,
	}

	if err := h.DeleteExpired(); err != nil {
		panic(err)
	}
}
