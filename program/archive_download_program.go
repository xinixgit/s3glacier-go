package program

import (
	"flag"
	"fmt"
	"os"
	"s3glacier-go/adapter"
	"s3glacier-go/domain"
	"s3glacier-go/svc"
	"time"
)

type ArchiveDownloadProgram struct {
	vault                string
	archiveId            string
	downloadId           int
	output               string
	chunkSizeInMB        int
	dbuser               string
	dbpwd                string
	dbname               string
	dbhost               string
	initialWaitTimeInHrs int
}

func (p *ArchiveDownloadProgram) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to download the archive from")
	fs.StringVar(&p.archiveId, "a", "", "The id of the archive to retrieve")
	fs.StringVar(&p.output, "o", "", "The output file")
	fs.StringVar(&p.dbuser, "u", "", "The username of the MySQL database")
	fs.StringVar(&p.dbpwd, "p", "", "The password of the MySQL database")
	fs.StringVar(&p.dbname, "db", "", "The name of the database created")
	fs.StringVar(&p.dbhost, "ip", "localhost", "The host name of the database, default to `localhost`")

	fs.IntVar(&p.chunkSizeInMB, "s", 1024, "The size of each chunk, defaults to 1GB (1024 * 1024 * 1024 bytes)")
	fs.IntVar(&p.initialWaitTimeInHrs, "w", 3, "Number of hours to wait before querying job status, default to 3 since S3 jobs are ready in 3-5 hrs")
	fs.IntVar(&p.downloadId, "dlID", -1, "The id of an existing download if a job has been created earlier")
}

func (p *ArchiveDownloadProgram) Run() error {
	s3g := createGlacierClient()
	csp := adapter.NewCloudServiceProvider(s3g)

	connStr := createConnStr(p.dbuser, p.dbpwd, p.dbhost, p.dbname)
	dao := adapter.NewDBDAO(connStr)

	sqsSvc := createSqsClient()
	notif := adapter.NewNotificationHandler(sqsSvc)

	dlSvc := svc.NewArchiveDownloadService(csp, dao, notif)
	file, err := createFileIfNecessary(p.output)
	if err != nil {
		return err
	}
	defer file.Close()

	notificationQueue := domain.NOTIF_QUEUE_NAME
	ctx := &svc.DownloadJobContext{
		ArchiveID:       &p.archiveId,
		Vault:           &p.vault,
		JobQueue:        &notificationQueue,
		ChunkSize:       p.chunkSizeInMB * domain.ONE_MB,
		InitialWaitTime: time.Duration(int64(p.initialWaitTimeInHrs) * int64(time.Hour)),
		WaitInterval:    domain.DefaultWaitInterval,
		File:            file,
	}

	if p.downloadId >= 0 {
		dlSvc.ResumeDownload(uint(p.downloadId), ctx)
		return nil
	}

	dlSvc.Download(ctx)
	return nil
}

func createFileIfNecessary(outputFile string) (*os.File, error) {
	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("unable to open output file %s: %w", outputFile, err)
	}

	return f, nil
}
