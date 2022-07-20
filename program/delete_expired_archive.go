package program

import (
	"flag"
	"s3glacier-go/adapter"
	"s3glacier-go/app"
)

type DeleteExpiredArchive struct {
	vault  string
	dbuser string
	dbpwd  string
	dbname string
	dbip   string
}

func (p *DeleteExpiredArchive) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to download the archive from")
	fs.StringVar(&p.dbuser, "u", "", "The username of the MySQL database")
	fs.StringVar(&p.dbpwd, "p", "", "The password of the MySQL database")
	fs.StringVar(&p.dbname, "db", "", "The name of the database created")
	fs.StringVar(&p.dbip, "ip", "localhost:3306", "The IP address and port number of the database, default to `localhost:3306`")
}

func (p *DeleteExpiredArchive) Run() {
	s3g := CreateGlacierClient()
	svc := adapter.NewCloudServiceProvider(s3g)

	connStr := CreateConnStr(p.dbuser, p.dbpwd, p.dbip, p.dbname)
	dao := adapter.NewDBDAO(connStr)

	repo := app.NewDeleteArchiveRepository(dao, svc)
	if err := repo.DeleteExpiredArchive(&p.vault); err != nil {
		panic(err)
	}
}
