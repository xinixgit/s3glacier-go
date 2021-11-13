package upload

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"xddd/s3glacier/db"
	"xddd/s3glacier/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

const (
	ONE_MB    = 1024 * 1024
	PART_SIZE = 1024 * ONE_MB // 1GB part size
)

type S3GlacierUploader struct {
	Vault     *string
	S3glacier *glacier.Glacier
	DBDAO     db.DBDAO
	Partno    int
}

func (u S3GlacierUploader) Upload(filePath string) {
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Unable to open file ", f)
		return
	}
	defer f.Close()

	fl := int(util.FileLen(f))
	fmt.Printf("Now start uploading file %s of %d bytes.\n", filePath, fl)

	uploadSessionId := u.initiateMultipartUpload()
	id := insertNewUpload(uploadSessionId, filePath, &u)

	finalChecksum := u.uploadSegments(uploadSessionId, filePath, fl, id, f)

	result := u.completeMultipartUpload(fl, finalChecksum, uploadSessionId)
	updateCompletedUpload(id, result, &u)

	fmt.Println("File uploaded, result: ", result)
}

func (u S3GlacierUploader) initiateMultipartUpload() *string {
	input := &glacier.InitiateMultipartUploadInput{
		AccountId:          aws.String("-"),
		PartSize:           aws.String(strconv.Itoa(PART_SIZE)),
		ArchiveDescription: aws.String("This is a test upload"),
		VaultName:          u.Vault,
	}

	out, err := u.S3glacier.InitiateMultipartUpload(input)
	if err != nil {
		panic(err)
	}

	fmt.Println("Multipart upload session created, id: ", *out.UploadId)
	return out.UploadId
}

func (u S3GlacierUploader) uploadSegments(uploadSessionId *string, filePath string, fl int, uploadId uint, f *os.File) *string {
	buf := make([]byte, PART_SIZE)
	hashes := [][]byte{}
	segCount := util.CeilQuotient(fl, PART_SIZE)
	segNum := 1
	off := 0

	if u.Partno > 1 {
		segNum = u.Partno
		off = (segNum - 1) * PART_SIZE
	}

	for off < fl {
		read, _ := f.ReadAt(buf, int64(off))
		seg := buf[:read]

		from := off
		to := off + read - 1
		checksum := ComputeSHA256TreeHash(seg, ONE_MB)
		hashes = append(hashes, checksum[:])
		r := fmt.Sprintf("bytes %d-%d/*", from, to)

		// upload a single segment in a multipart upload session
		input := &glacier.UploadMultipartPartInput{
			AccountId: aws.String("-"),
			Body:      bytes.NewReader(seg),
			Checksum:  aws.String(hex.EncodeToString(checksum)),
			Range:     aws.String(r),
			UploadId:  uploadSessionId,
			VaultName: u.Vault,
		}

		result, err := u.S3glacier.UploadMultipartPart(input)
		if err != nil {
			fmt.Printf("(%d/%d) failed for upload %d with file %s.\n", segNum, segCount, uploadId, filePath)
			updateFailedUpload(uploadId, &u)
			panic(err)
		}

		fmt.Printf("(%d/%d) uploaded for upload %d.\n", segNum, segCount, uploadId)
		insertUploadedSegment(result, segNum, segCount, uploadId, &u)

		segNum = segNum + 1
		off = off + read
	}

	encoded := hex.EncodeToString(ComputeCombineHashChunks(hashes))
	return &encoded
}

func (u S3GlacierUploader) abortMultipartUpload(uploadSessionId *string) {
	input := &glacier.AbortMultipartUploadInput{
		AccountId: aws.String("-"),
		UploadId:  uploadSessionId,
		VaultName: u.Vault,
	}

	_, err := u.S3glacier.AbortMultipartUpload(input)
	if err != nil {
		fmt.Println("Abort multipart upload failed for ", *uploadSessionId)
		fmt.Println(err)
	}
}

func (u S3GlacierUploader) completeMultipartUpload(fl int, checksum *string, uploadSessionId *string) *glacier.ArchiveCreationOutput {
	input := &glacier.CompleteMultipartUploadInput{
		AccountId:   aws.String("-"),
		ArchiveSize: aws.String(strconv.Itoa(fl)),
		Checksum:    checksum,
		UploadId:    uploadSessionId,
		VaultName:   u.Vault,
	}
	res, err := u.S3glacier.CompleteMultipartUpload(input)
	if err != nil {
		panic(err)
	}
	return res
}
