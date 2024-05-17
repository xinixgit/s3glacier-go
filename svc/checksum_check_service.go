package svc

import (
	"os"
	"s3glacier-go/util"
)

type checksumCheckService struct{}

func NewChecksumCheckService() *checksumCheckService {
	return &checksumCheckService{}
}

func (s *checksumCheckService) GenerateChecksum(file string, offset int64) string {
	f, err := os.Open(file)
	panicIfErr(err)

	stat, err := f.Stat()
	panicIfErr(err)

	buf := make([]byte, stat.Size()-offset)
	f.ReadAt(buf, offset)

	return toHexString(util.ComputeSHA256TreeHashWithOneMBChunks(buf))
}
