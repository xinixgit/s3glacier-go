package util

import (
	"bytes"
	"crypto/sha256"
	"io"
	"s3glacier-go/domain"

	"github.com/aws/aws-sdk-go/service/glacier"
)

func ComputeSHA256TreeHashWithOneMBChunks(data []byte) []byte {
	return ComputeSHA256TreeHash(data, domain.ONE_MB)
}

// Splits data into {chunkSize} chunks, and calculate the final hash by
// combining the hash of each chunk
func ComputeSHA256TreeHash(data []byte, chunkSize int) []byte {
	hashes := getHashesChunks(data, chunkSize)
	return ComputeCombineHashChunks(hashes)
}

func getHashesChunks(data []byte, chunkSize int) [][]byte {
	r := bytes.NewReader(data)
	buf := make([]byte, chunkSize)
	hashes := [][]byte{}

	for {
		n, err := io.ReadAtLeast(r, buf, chunkSize)
		if n == 0 {
			break
		}

		tmpHash := sha256.Sum256(buf[:n])
		hashes = append(hashes, tmpHash[:])
		if err != nil {
			break
		}
	}
	return hashes
}

func ComputeCombineHashChunks(hashes [][]byte) []byte {
	return glacier.ComputeTreeHash(hashes)
}
