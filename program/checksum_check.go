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
	fs.StringVar(&c.filePath, "f", "", "the path of the file to calculate checksum against")
	fs.Int64Var(&c.offset, "o", 0, "the offset in bytes to read the file from")
	fs.StringVar(&c.expected, "e", "", "the expected checksum")
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

	if c.expected != "" {
		fmt.Println("Match with expected value: ", c.expected == encoded)
	}
}
