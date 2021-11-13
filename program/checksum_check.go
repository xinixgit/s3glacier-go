package program

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"xddd/s3glacier/upload"
	"xddd/s3glacier/util"
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
	f, err := os.Open(c.filePath)
	util.PanicIfErr(err)

	stat, err := f.Stat()
	util.PanicIfErr(err)

	buf := make([]byte, stat.Size()-c.offset)
	f.ReadAt(buf, c.offset)

	checksum := upload.ComputeSHA256TreeHash(buf, upload.ONE_MB)
	encoded := hex.EncodeToString(checksum)
	fmt.Println("Checksum: ", encoded)

	if c.expected != "EMPTY" {
		fmt.Println("Match with expected value: ", c.expected == encoded)
	}
}
