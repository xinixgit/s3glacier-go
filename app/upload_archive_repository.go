package app

import (
	"fmt"
	"os"
	"s3glacier-go/domain"
	"s3glacier-go/util"
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

func (ctx *UploadJobContext) hasResumedUpload() bool {
	return ctx.UploadID > 0
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

	finalChecksum, err := repo.uploadSegments(uploadSessionId, id, filename, fl, ctx)
	if err != nil {
		return err
	}

	res, err := repo.svc.CompleteMultipartUploadInput(fl, finalChecksum, uploadSessionId, ctx.Vault)
	if err != nil {
		return err
	}

	repo.updateCompletedUpload(id, res)

	fmt.Println("File uploaded, result: ", res)
	return nil
}

func (repo *UploadArchiveRepositoryImpl) uploadSegments(uploadSessionId *string, uploadId uint, filename string, fl int64, ctx *UploadJobContext) (*string, error) {
	buf := make([]byte, ctx.ChunkSize)
	hashes := [][]byte{}
	segCount := int(util.CeilQuotient(fl, int64(ctx.ChunkSize)))
	off, segNum := int64(0), 1

	for off < fl {
		read, _ := ctx.File.ReadAt(buf, off)
		seg := buf[:read]

		from := off
		to := off + int64(read) - 1
		checksum := util.ComputeSHA256TreeHashWithOneMBChunks(seg)
		hashes = append(hashes, checksum[:])

		// If we are resuming from a previously failed upload, we do not need to run the upload if the segment is already
		// uploaded (but still need to calculate the hash of previous segments).
		if !ctx.hasResumedUpload() || (segNum > int(ctx.MaxSegNum)) {
			byteRange := util.GetBytesRangeInt64(from, to)
			checksum := util.ToHexString(checksum)

			_, err := repo.svc.UploadMultipartPart(seg, &checksum, &byteRange, uploadSessionId, ctx.Vault)
			if err != nil {
				fmt.Printf("(%d/%d) failed for upload id %d with file %s.\n", segNum, segCount, uploadId, filename)
				repo.updateFailedUpload(uploadId)
				if err1 := repo.svc.AbortMultipartUpload(uploadSessionId, ctx.Vault); err1 != nil {
					fmt.Println("Unable to abort upload.", err1)
				}
				return nil, err
			}

			fmt.Printf("(%d/%d) with %s has been uploaded for upload id %d.\n", segNum, segCount, byteRange, uploadId)
			repo.insertUploadedSegment(&checksum, segNum, segCount, uploadId)
		}

		segNum = segNum + 1
		off = off + int64(read)
	}

	encoded := util.ToHexString(util.ComputeCombineHashChunks(hashes))
	return &encoded, nil
}
