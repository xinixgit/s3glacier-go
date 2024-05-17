package program

import (
	"flag"
	"s3glacier-go/adapter"
	"s3glacier-go/svc"
)

type ArchiveDeleteProgram struct {
	vault  string
	dbuser string
	dbpwd  string
	dbname string
	dbhost string
}

func (p *ArchiveDeleteProgram) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to download the archive from")
	fs.StringVar(&p.dbuser, "u", "", "The username of the MySQL database")
	fs.StringVar(&p.dbpwd, "p", "", "The password of the MySQL database")
	fs.StringVar(&p.dbname, "db", "", "The name of the database created")
	fs.StringVar(&p.dbhost, "ip", "localhost", "The host name of the database, default to `localhost`")
}

func (p *ArchiveDeleteProgram) Run() {
	s3g := createGlacierClient()
	csp := adapter.NewCloudServiceProvider(s3g)

	connStr := createConnStr(p.dbuser, p.dbpwd, p.dbhost, p.dbname)
	dao := adapter.NewDBDAO(connStr)

	delSvc := svc.NewArchiveDeleteService(dao, csp)
	if err := delSvc.DeleteExpiredArchive(&p.vault); err != nil {
		panic(err)
	}
}
