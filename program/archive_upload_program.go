package program

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"s3glacier-go/adapter"
	"s3glacier-go/domain"
	svc "s3glacier-go/svc"
)

type ArchiveUploadProgram struct {
	vault         string
	fpat          string
	chunkSizeInMB int
	dbuser        string
	dbpwd         string
	dbname        string
	dbhost        string
	uploadId      uint
}

func (p *ArchiveUploadProgram) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.vault, "v", "", "The name of the vault to upload the archive to")
	fs.StringVar(&p.fpat, "f", "", "The regex of the archive files to be uploaded, you can use `*` to upload all files in a folder, or specify a single file")
	fs.StringVar(&p.dbuser, "u", "", "The username of the MySQL database")
	fs.StringVar(&p.dbpwd, "p", "", "The password of the MySQL database")
	fs.StringVar(&p.dbname, "db", "", "The name of the database created")
	fs.StringVar(&p.dbhost, "ip", "localhost", "The host name of the database, default to `localhost`")

	fs.IntVar(&p.chunkSizeInMB, "s", 1024, "The size of each chunk, defaults to 1GB (1024 * 1024 * 1024 bytes)")
	fs.UintVar(&p.uploadId, "uploadId", 0, "The id of the upload (from the `uploads` table) to resume, if some of its parts had failed to be uploaded previously")
}

func (p *ArchiveUploadProgram) Run() error {
	files, err := filepath.Glob(p.fpat)
	if err != nil {
		return fmt.Errorf("read file failed: %w", err)
	}

	if len(files) > 1 && p.uploadId > 0 {
		return fmt.Errorf("upload id only applies when uploading a single file")
	}

	if p.chunkSizeInMB%2 != 0 {
		return fmt.Errorf("chunk size has to be the power of 2")
	}

	csp := adapter.NewCloudServiceProvider(createGlacierClient())
	dao := adapter.NewDBDAO(createConnStr(p.dbuser, p.dbpwd, p.dbhost, p.dbname))
	uplSvc := svc.NewArchiveUploadService(csp, dao)
	ctx := svc.UploadJobContext{
		UploadID:  p.uploadId,
		Vault:     &p.vault,
		ChunkSize: p.chunkSizeInMB * domain.ONE_MB,
	}

	if ctx.HasResumedUpload() {
		ctx.MaxSegNum = dao.GetMaxSegNumByUploadID(p.uploadId)
	}

	for _, filePath := range files {
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("unable to open file path %s: %w", filePath, err)
		}
		defer f.Close()

		ctx.File = f
		if err := uplSvc.Upload(&ctx); err != nil {
			return err
		}
	}

	return nil
}
