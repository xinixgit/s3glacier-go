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
	ONE_MB = 1024 * 1024
)

type ResumedUpload struct {
	Upload    *db.Upload
	MaxSegNum int
}

type S3GlacierUploader struct {
	Vault         *string
	ChunkSize     int
	S3glacier     *glacier.Glacier
	DBDAO         db.DBDAO
	ResumedUpload *ResumedUpload
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

	var (
		id              uint
		uploadSessionId *string
	)
	if u.ResumedUpload != nil {
		uploadSessionId = &u.ResumedUpload.Upload.SessionId
		id = u.ResumedUpload.Upload.ID
	} else {
		uploadSessionId = u.initiateMultipartUpload()
		id = insertNewUpload(uploadSessionId, filePath, &u)
	}

	finalChecksum := u.uploadSegments(uploadSessionId, filePath, fl, id, f)

	result := u.completeMultipartUpload(fl, finalChecksum, uploadSessionId)
	updateCompletedUpload(id, result, &u)

	fmt.Println("File uploaded, result: ", result)
}

func (u S3GlacierUploader) initiateMultipartUpload() *string {
	input := &glacier.InitiateMultipartUploadInput{
		AccountId:          aws.String("-"),
		PartSize:           aws.String(strconv.Itoa(u.ChunkSize)),
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
	buf := make([]byte, u.ChunkSize)
	hashes := [][]byte{}
	segCount := util.CeilQuotient(fl, u.ChunkSize)
	off, segNum := 0, 1

	for off < fl {
		read, _ := f.ReadAt(buf, int64(off))
		seg := buf[:read]

		from := off
		to := off + read - 1
		checksum := ComputeSHA256TreeHash(seg, ONE_MB)
		hashes = append(hashes, checksum[:])

		// If we are resuming from a previously failed upload, we do not need to run the upload if the segment is already
		// uploaded (but still need to calculate the hash of previous segments).
		if u.ResumedUpload == nil || segNum > u.ResumedUpload.MaxSegNum {
			r := fmt.Sprintf("bytes %d-%d/*", from, to)
			input := &glacier.UploadMultipartPartInput{
				AccountId: aws.String("-"),
				Body:      bytes.NewReader(seg),
				Checksum:  aws.String(hex.EncodeToString(checksum)),
				Range:     aws.String(r),
				UploadId:  uploadSessionId,
				VaultName: u.Vault,
			}

			// upload a single segment in a multipart upload session
			result, err := u.S3glacier.UploadMultipartPart(input)
			if err != nil {
				fmt.Printf("(%d/%d) failed for upload id %d with file %s.\n", segNum, segCount, uploadId, filePath)
				updateFailedUpload(uploadId, &u)
				panic(err)
			}

			fmt.Printf("(%d/%d) with bytes range %s has been uploaded for upload id %d.\n", segNum, segCount, r, uploadId)
			insertUploadedSegment(result, segNum, segCount, uploadId, &u)
		}

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
