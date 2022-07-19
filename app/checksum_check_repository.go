package app

import (
	"os"
	"s3glacier-go/util"
)

type ChecksumCheckRepository interface {
	GenerateChecksum(file string, offset int64) string
}

type ChecksumCheckRepositoryImpl struct {
}

func NewChecksumCheckRepository() ChecksumCheckRepository {
	return &ChecksumCheckRepositoryImpl{}
}

func (c ChecksumCheckRepositoryImpl) GenerateChecksum(file string, offset int64) string {
	f, err := os.Open(file)
	util.PanicIfErr(err)

	stat, err := f.Stat()
	util.PanicIfErr(err)

	buf := make([]byte, stat.Size()-offset)
	f.ReadAt(buf, offset)

	return util.ToHexString(util.ComputeSHA256TreeHashWithOneMBChunks(buf))
}
