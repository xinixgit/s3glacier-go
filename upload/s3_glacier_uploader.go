package upload

import (
	"bytes"
	"fmt"
	"os"
	"s3glacier-go/db"
	"s3glacier-go/util"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
)

type ResumedUpload struct {
	Upload    *db.Upload
	MaxSegNum int
}

type S3GlacierUploader struct {
	Vault         *string
	ChunkSize     int
	S3glacier     *glacier.Glacier
	UlDAO         db.UploadDAO
	ResumedUpload *ResumedUpload
}

func (ul S3GlacierUploader) Upload(f *os.File) {
	filename := f.Name()
	fl := int(util.FileLen(f))
	fmt.Printf("Now start uploading file %s of %d bytes.\n", filename, fl)

	var (
		id              uint
		uploadSessionId *string
	)
	if ul.ResumedUpload != nil {
		uploadSessionId = &ul.ResumedUpload.Upload.SessionId
		id = ul.ResumedUpload.Upload.ID
	} else {
		uploadSessionId = ul.initiateMultipartUpload()
		id = insertNewUpload(uploadSessionId, filename, ul.Vault, ul.UlDAO)
	}

	finalChecksum := ul.uploadSegments(uploadSessionId, filename, fl, id, f)

	result := ul.completeMultipartUpload(fl, finalChecksum, uploadSessionId)
	updateCompletedUpload(id, result, ul.UlDAO)

	fmt.Println("File uploaded, result: ", result)
}

func (u S3GlacierUploader) initiateMultipartUpload() *string {
	input := &glacier.InitiateMultipartUploadInput{
		AccountId: aws.String("-"),
		PartSize:  aws.String(strconv.Itoa(u.ChunkSize)),
		VaultName: u.Vault,
	}

	out, err := u.S3glacier.InitiateMultipartUpload(input)
	if err != nil {
		panic(err)
	}

	fmt.Println("Multipart upload session created, id: ", *out.UploadId)
	return out.UploadId
}

func (ul S3GlacierUploader) uploadSegments(uploadSessionId *string, filename string, fl int, uploadId uint, f *os.File) *string {
	buf := make([]byte, ul.ChunkSize)
	hashes := [][]byte{}
	segCount := util.CeilQuotient(fl, ul.ChunkSize)
	off, segNum := 0, 1

	for off < fl {
		read, _ := f.ReadAt(buf, int64(off))
		seg := buf[:read]

		from := off
		to := off + read - 1
		checksum := util.ComputeSHA256TreeHashWithOneMBChunks(seg)
		hashes = append(hashes, checksum[:])

		// If we are resuming from a previously failed upload, we do not need to run the upload if the segment is already
		// uploaded (but still need to calculate the hash of previous segments).
		if ul.ResumedUpload == nil || segNum > ul.ResumedUpload.MaxSegNum {
			r := util.GetBytesRange(from, to)
			input := &glacier.UploadMultipartPartInput{
				AccountId: aws.String("-"),
				Body:      bytes.NewReader(seg),
				Checksum:  aws.String(util.ToHexString(checksum)),
				Range:     aws.String(r),
				UploadId:  uploadSessionId,
				VaultName: ul.Vault,
			}

			// upload a single segment in a multipart upload session
			result, err := ul.S3glacier.UploadMultipartPart(input)
			if err != nil {
				fmt.Printf("(%d/%d) failed for upload id %d with file %s.\n", segNum, segCount, uploadId, filename)
				updateFailedUpload(uploadId, ul.UlDAO)
				panic(err)
			}

			fmt.Printf("(%d/%d) with %s has been uploaded for upload id %d.\n", segNum, segCount, r, uploadId)
			insertUploadedSegment(result, segNum, segCount, uploadId, ul.UlDAO)
		}

		segNum = segNum + 1
		off = off + read
	}

	encoded := util.ToHexString(util.ComputeCombineHashChunks(hashes))
	return &encoded
}

func (ul S3GlacierUploader) abortMultipartUpload(uploadSessionId *string) {
	input := &glacier.AbortMultipartUploadInput{
		AccountId: aws.String("-"),
		UploadId:  uploadSessionId,
		VaultName: ul.Vault,
	}

	if _, err := ul.S3glacier.AbortMultipartUpload(input); err != nil {
		fmt.Println("Abort multipart upload failed for ", *uploadSessionId)
		fmt.Println(err)
	}
}

func (ul S3GlacierUploader) completeMultipartUpload(fl int, checksum *string, uploadSessionId *string) *glacier.ArchiveCreationOutput {
	input := &glacier.CompleteMultipartUploadInput{
		AccountId:   aws.String("-"),
		ArchiveSize: aws.String(strconv.Itoa(fl)),
		Checksum:    checksum,
		UploadId:    uploadSessionId,
		VaultName:   ul.Vault,
	}
	res, err := ul.S3glacier.CompleteMultipartUpload(input)
	if err != nil {
		panic(err)
	}
	return res
}
