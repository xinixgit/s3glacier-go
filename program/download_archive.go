package program

import (
	"flag"
	"os"
	"s3glacier-go/adapter"
	"s3glacier-go/app"
	"s3glacier-go/domain"
	"time"
)

type DownloadArchive struct {
	vault                string
	archiveId            string
	downloadId           int
	output               string
	chunkSizeInMB        int
	dbuser               string
	dbpwd                string
	dbname               string
	dbip                 string
	initialWaitTimeInHrs int
}

func (p *DownloadArchive) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to download the archive from")
	fs.StringVar(&p.archiveId, "a", "", "The id of the archive to retrieve")
	fs.StringVar(&p.output, "o", "", "The output file")
	fs.StringVar(&p.dbuser, "u", "", "The username of the MySQL database")
	fs.StringVar(&p.dbpwd, "p", "", "The password of the MySQL database")
	fs.StringVar(&p.dbname, "db", "", "The name of the database created")
	fs.StringVar(&p.dbip, "ip", "localhost:3306", "The IP address and port number of the database, default to `localhost:3306`")

	fs.IntVar(&p.chunkSizeInMB, "s", 1024, "The size of each chunk, defaults to 1GB (1024 * 1024 * 1024 bytes)")
	fs.IntVar(&p.initialWaitTimeInHrs, "w", 3, "Number of hours to wait before querying job status, default to 3 since S3 jobs are ready in 3-5 hrs")
	fs.IntVar(&p.downloadId, "dlID", -1, "The id of an existing download if a job has been created earlier")
}

func (p *DownloadArchive) Run() {
	s3g := CreateGlacierClient()
	svc := adapter.NewCloudServiceProvider(s3g)

	connStr := CreateConnStr(p.dbuser, p.dbpwd, p.dbip, p.dbname)
	dao := adapter.NewDBDAO(connStr)

	repo := app.NewDownloadArchiveRepository(svc, dao)
	file := createFileIfNecessary(p.output)
	defer file.Close()

	ctx := &app.DownloadJobContext{
		ArchiveID:       &p.archiveId,
		Vault:           &p.vault,
		ChunkSize:       p.chunkSizeInMB * domain.ONE_MB,
		InitialWaitTime: time.Duration(int64(p.initialWaitTimeInHrs) * int64(time.Hour)),
		WaitInterval:    domain.DefaultWaitInterval,
		File:            file,
	}

	if p.downloadId >= 0 {
		repo.ResumeDownload(uint(p.downloadId), ctx)
		return
	}

	repo.Download(ctx)
}

func createFileIfNecessary(outputFile string) (f *os.File) {
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return
}
