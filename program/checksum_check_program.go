package program

import (
	"flag"
	"fmt"
	"s3glacier-go/svc"
)

type ChecksumCheckProgram struct {
	filePath string
	offset   int64
	expected string
}

func (p *ChecksumCheckProgram) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&p.filePath, "f", "", "The regex of the archive files to be uploaded, you can use `*` to upload all files in a folder, or specify a single file")
	fs.Int64Var(&p.offset, "o", 0, "The offset in bytes to read the file with, defaults to 0 (start of the file)")
	fs.StringVar(&p.expected, "e", "EMPTY", "The expected checksum, defaults to empty")
}

func (p *ChecksumCheckProgram) Run() {
	chkSvc := svc.NewChecksumCheckService()
	checksum := chkSvc.GenerateChecksum(p.filePath, p.offset)
	fmt.Println("Checksum: ", checksum)

	if p.expected != "EMPTY" {
		fmt.Println("Match with expected value: ", p.expected == checksum)
	}
}
