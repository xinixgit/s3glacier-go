package app

import (
	"fmt"
	"os"
	"s3glacier-go/domain"
	"s3glacier-go/util"

	"github.com/aws/aws-sdk-go/service/glacier"
)

type UploadArchiveRepository interface {
	Upload(ctx *UploadJobContext) error
}

type UploadArchiveRepositoryImpl struct {
	svc domain.CloudServiceProvider
	dao domain.DBDAO
}

func NewUploadArchiveRepository(svc domain.CloudServiceProvider, dao domain.DBDAO) UploadArchiveRepository {
	return &UploadArchiveRepositoryImpl{
		svc: svc,
		dao: dao,
	}
}

type UploadJobContext struct {
	UploadID  uint
	MaxSegNum uint
	Vault     *string
	ChunkSize int
	File      *os.File
}

func (repo *UploadArchiveRepositoryImpl) Upload(ctx *UploadJobContext) error {
	filename := ctx.File.Name()
	fl := util.FileLen(ctx.File)
	fmt.Printf("Now start uploading file %s of %d bytes.\n", filename, fl)

	var (
		id              uint
		uploadSessionId *string
	)

	if ctx.UploadID > 0 {
		upload := repo.dao.GetUploadByID(ctx.UploadID)
		uploadSessionId = &upload.SessionId
		id = upload.ID
	} else {
		uploadSessionId, err := repo.svc.InitiateMultipartUpload(ctx.ChunkSize, ctx.Vault)
		if err != nil {
			return err
		}
		fmt.Println("Multipart upload session created, id: ", *uploadSessionId)
		id = repo.insertNewUpload(uploadSessionId, filename, ctx.Vault)
	}

	finalChecksum := repo.uploadSegments(uploadSessionId, id, filename, fl, ctx.File)

	res := repo.completeMultipartUpload(fl, finalChecksum, uploadSessionId)
	repo.updateCompletedUpload(id, &domain.ArchiveCreationOutput{
		Location:  res.Location,
		Checksum:  res.Checksum,
		ArchiveId: res.ArchiveId,
	})

	fmt.Println("File uploaded, result: ", res)
	return nil
}

func (repo *UploadArchiveRepositoryImpl) uploadSegments(uploadSessionId *string, uploadId uint, filename string, fl int64, f *os.File) *string {
	// buf := make([]byte, ul.ChunkSize)
	// hashes := [][]byte{}
	// segCount := int(util.CeilQuotient(fl, int64(ul.ChunkSize)))
	// off, segNum := int64(0), 1

	// for off < fl {
	// 	read, _ := f.ReadAt(buf, off)
	// 	seg := buf[:read]

	// 	from := off
	// 	to := off + int64(read) - 1
	// 	checksum := util.ComputeSHA256TreeHashWithOneMBChunks(seg)
	// 	hashes = append(hashes, checksum[:])

	// 	// If we are resuming from a previously failed upload, we do not need to run the upload if the segment is already
	// 	// uploaded (but still need to calculate the hash of previous segments).
	// 	if ul.ResumedUpload == nil || segNum > ul.ResumedUpload.MaxSegNum {
	// 		r := util.GetBytesRangeInt64(from, to)
	// 		input := &glacier.UploadMultipartPartInput{
	// 			AccountId: aws.String("-"),
	// 			Body:      bytes.NewReader(seg),
	// 			Checksum:  aws.String(util.ToHexString(checksum)),
	// 			Range:     aws.String(r),
	// 			UploadId:  uploadSessionId,
	// 			VaultName: ul.Vault,
	// 		}

	// 		// upload a single segment in a multipart upload session
	// 		result, err := ul.S3glacier.UploadMultipartPart(input)
	// 		if err != nil {
	// 			fmt.Printf("(%d/%d) failed for upload id %d with file %s.\n", segNum, segCount, uploadId, filename)
	// 			updateFailedUpload(uploadId, ul.UlDAO)
	// 			panic(err)
	// 		}

	// 		fmt.Printf("(%d/%d) with %s has been uploaded for upload id %d.\n", segNum, segCount, r, uploadId)
	// 		insertUploadedSegment(result, segNum, segCount, uploadId, ul.UlDAO)
	// 	}

	// 	segNum = segNum + 1
	// 	off = off + int64(read)
	// }

	// encoded := util.ToHexString(util.ComputeCombineHashChunks(hashes))
	// return &encoded
	return nil
}

func (repo *UploadArchiveRepositoryImpl) abortMultipartUpload(uploadSessionId *string) {
	// input := &glacier.AbortMultipartUploadInput{
	// 	AccountId: aws.String("-"),
	// 	UploadId:  uploadSessionId,
	// 	VaultName: ul.Vault,
	// }

	// if _, err := ul.S3glacier.AbortMultipartUpload(input); err != nil {
	// 	fmt.Println("Abort multipart upload failed for ", *uploadSessionId)
	// 	fmt.Println(err)
	// }
}

func (repo *UploadArchiveRepositoryImpl) completeMultipartUpload(fl int64, checksum *string, uploadSessionId *string) *glacier.ArchiveCreationOutput {
	// input := &glacier.CompleteMultipartUploadInput{
	// 	AccountId:   aws.String("-"),
	// 	ArchiveSize: aws.String(strconv.FormatInt(fl, 10)),
	// 	Checksum:    checksum,
	// 	UploadId:    uploadSessionId,
	// 	VaultName:   ul.Vault,
	// }
	// res, err := ul.S3glacier.CompleteMultipartUpload(input)
	// if err != nil {
	// 	panic(err)
	// }
	// return res
	return nil
}
