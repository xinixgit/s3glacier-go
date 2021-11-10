package upload

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"strconv"
	"encoding/hex"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glacier"
	"xddd/s3glacier/util"
	"xddd/s3glacier/db"
)

var ONE_MB = 1024 * 1024
var PART_SIZE = 256 * ONE_MB

type S3GlacierUploader struct {
	Vault 			*string
	S3glacier 	*glacier.Glacier
	DBDAO 			db.DBDAO
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

	uploadSessionId := u.InitiateMultipartUpload()
	id := u.InsertNewUpload(uploadSessionId, filePath)

	finalChecksum := u.UploadSegments(uploadSessionId, filePath, fl, id, f)

	result := u.CompleteMultipartUpload(fl, finalChecksum, uploadSessionId)
	u.UpdateCompletedUpload(id, result)

	fmt.Println("File uploaded, result: ", result)
}

func (u S3GlacierUploader) InitiateMultipartUpload() *string {
	input := &glacier.InitiateMultipartUploadInput{
    AccountId: 					aws.String("-"),
    PartSize:  					aws.String(strconv.Itoa(PART_SIZE)),
    ArchiveDescription: aws.String("This is a test upload"),
    VaultName: 					u.Vault,
	}

	out, err := u.S3glacier.InitiateMultipartUpload(input)
	if err != nil {
		panic(err)
	}

	fmt.Println("Multipart upload session created, id: ", *out.UploadId)
	return out.UploadId
}


func (u S3GlacierUploader) UploadSegments(uploadSessionId *string, filePath string, fl int, uploadId uint, f *os.File) *string {
	off := 0
	buf := make([]byte, PART_SIZE)
	hashes := [][]byte{}
	segNum := 1
	segCount := util.CeilQuotient(fl, PART_SIZE)

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
			fmt.Printf("(%d/%d) failed for file %s.\n", segNum, segCount, filePath)
			u.AbortMultipartUpload(uploadSessionId)
			panic(err)
		}

		fmt.Printf("(%d/%d) uploaded for upload %d.\n", segNum, segCount, uploadId)
		u.InsertUploadedSegment(result, segNum, segCount, uploadId)

		segNum = segNum + 1
		off = off + read
	}

	encoded := hex.EncodeToString(ComputeCombineHashChunks(hashes))
	return &encoded
}

func (u S3GlacierUploader) AbortMultipartUpload(uploadSessionId *string) {
	input := &glacier.AbortMultipartUploadInput{
		AccountId:  	aws.String("-"),
		UploadId:   	uploadSessionId,
		VaultName:		u.Vault,
	}

	_, err := u.S3glacier.AbortMultipartUpload(input)
	if err != nil {
		fmt.Println("Abort multipart upload failed for ", *uploadSessionId)
		fmt.Println(err)
	}
}

func (u S3GlacierUploader) InsertUploadedSegment(result *glacier.UploadMultipartPartOutput, segNum int, segCount int, uploadId uint) {
	if uploadId == 0 {
		return
	}

	seg := &db.UploadedSegment{
		SegmentNum:  	segNum,
		UploadId: 		uploadId,
		Checksum: 		*result.Checksum,
		CreatedAt:  	time.Now().Format("2006-01-02 15:04:05"),
	}

	err := u.DBDAO.InsertUploadedSegment(seg)
	if err != nil {
		fmt.Printf("Insert uploaded segment failed for seg num %d, and upload id %d.\n", segNum, uploadId)
	}
}

func (u S3GlacierUploader) CompleteMultipartUpload(fl int, checksum *string, uploadSessionId *string) *glacier.ArchiveCreationOutput {
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

func (u S3GlacierUploader) InsertNewUpload(sessionId *string, filename string) uint {
	upload := &db.Upload{
		VaultName: 	*u.Vault,
		Filename:		filename,
		SessionId:	*sessionId,
		CreatedAt:  time.Now().Format("2006-01-02 15:04:05"),
		Status:  		db.STARTED,
	}

	err := u.DBDAO.InsertUpload(upload)
	if err != nil {
		fmt.Printf("Insert upload failed for %s and session %s.\n", *sessionId, filename)
	}
	return upload.ID
}

func (u S3GlacierUploader) UpdateCompletedUpload(id uint, res *glacier.ArchiveCreationOutput) {
	upload := u.DBDAO.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("Failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Location 	= *res.Location
	upload.Checksum 	= *res.Checksum
	upload.ArchiveId 	= *res.ArchiveId
	upload.Status 		= db.COMPLETED
	u.DBDAO.UpdateUpload(upload)
}
