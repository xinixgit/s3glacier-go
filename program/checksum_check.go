package program

import (
	"flag"
	"fmt"
	"s3glacier-go/app"
)

type ChecksumCheck struct {
	filePath string
	offset   int64
	expected string
}

func (c *ChecksumCheck) InitFlag(fs *flag.FlagSet) {
	fs.StringVar(&c.filePath, "f", "", "The regex of the archive files to be uploaded, you can use `*` to upload all files in a folder, or specify a single file")
	fs.Int64Var(&c.offset, "o", 0, "The offset in bytes to read the file with, defaults to 0 (start of the file)")
	fs.StringVar(&c.expected, "e", "EMPTY", "The expected checksum, defaults to empty")
}

func (c *ChecksumCheck) Run() {
	check := app.NewChecksumCheckRepository()
	checksum := check.GenerateChecksum(c.filePath, c.offset)
	fmt.Println("Checksum: ", checksum)

	if c.expected != "EMPTY" {
		fmt.Println("Match with expected value: ", c.expected == checksum)
	}
}
