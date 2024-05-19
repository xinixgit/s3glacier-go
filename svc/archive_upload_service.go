package svc

import (
	"fmt"
	"os"
	"s3glacier-go/domain"
	"s3glacier-go/util"
)

type UploadJobContext struct {
	UploadID  uint
	MaxSegNum int
	Vault     *string
	ChunkSize int
	File      *os.File
}

func (ctx *UploadJobContext) HasResumedUpload() bool {
	return ctx.UploadID > 0
}

type archiveUploadService interface {
	Upload(ctx *UploadJobContext) error
}

type archiveUploadServiceImpl struct {
	csp domain.CloudServiceProvider
	dao domain.DBDAO
}

func NewArchiveUploadService(scp domain.CloudServiceProvider, dao domain.DBDAO) archiveUploadService {
	return &archiveUploadServiceImpl{
		csp: scp,
		dao: dao,
	}
}

func (s *archiveUploadServiceImpl) Upload(ctx *UploadJobContext) error {
	filename := ctx.File.Name()
	fl := fileLen(ctx.File)
	fmt.Printf("Now start uploading file %s of %d bytes.\n", filename, fl)

	var (
		id              uint
		uploadSessionId *string
	)

	if ctx.HasResumedUpload() {
		upload := s.dao.GetUploadByID(ctx.UploadID)
		uploadSessionId = &upload.SessionId
		id = upload.ID
	} else {
		var err error
		uploadSessionId, err = s.csp.InitiateMultipartUpload(ctx.ChunkSize, ctx.Vault)
		if err != nil {
			return err
		}
		fmt.Printf("Multipart upload session created, id: %s\n", *uploadSessionId)
		id = s.insertNewUpload(uploadSessionId, filename, ctx.Vault)
	}

	finalChecksum, err := s.uploadSegments(uploadSessionId, id, filename, fl, ctx)
	if err != nil {
		return err
	}

	res, err := s.csp.CompleteMultipartUploadInput(fl, finalChecksum, uploadSessionId, ctx.Vault)
	if err != nil {
		return err
	}

	s.updateCompletedUpload(id, res)

	fmt.Println("File uploaded, result: ", res)
	return nil
}

func (s *archiveUploadServiceImpl) uploadSegments(
	uploadSessionId *string,
	uploadId uint,
	filename string,
	fl int64,
	ctx *UploadJobContext,
) (*string, error) {
	buf := make([]byte, ctx.ChunkSize)
	hashes := [][]byte{}
	segCount := int(ceilQuotient(fl, int64(ctx.ChunkSize)))
	offset, segNum := int64(0), 1

	for offset < fl {
		read, _ := ctx.File.ReadAt(buf, offset)
		seg := buf[:read]

		from := offset
		to := offset + int64(read) - 1
		checksum := util.ComputeSHA256TreeHashWithOneMBChunks(seg)
		hashes = append(hashes, checksum[:])

		// If we are resuming from a previously failed upload, we do not need to run the upload if the segment
		// is already uploaded (but still need to calculate the hash of previous segments).
		if !ctx.HasResumedUpload() || (segNum > ctx.MaxSegNum) {
			byteRange := getBytesRangeInt64(from, to)
			checksum := toHexString(checksum)

			_, err := s.csp.UploadMultipartPart(seg, &checksum, &byteRange, uploadSessionId, ctx.Vault)
			if err != nil {
				s.updateFailedUpload(uploadId)
				if err1 := s.csp.AbortMultipartUpload(uploadSessionId, ctx.Vault); err1 != nil {
					fmt.Printf("unable to abort upload session: %s\n", err1)
				}

				return nil, fmt.Errorf(
					"segment (%d/%d) upload failed for upload id %d with file %s: %w",
					segNum,
					segCount,
					uploadId,
					filename,
					err,
				)
			}

			fmt.Printf("(%d/%d) with %s has been uploaded for upload id %d.\n", segNum, segCount, byteRange, uploadId)
			s.insertUploadedSegment(&checksum, segNum, segCount, uploadId)
		}

		segNum += 1
		offset += int64(read)
	}

	encoded := toHexString(util.ComputeCombineHashChunks(hashes))
	return &encoded, nil
}

func (s *archiveUploadServiceImpl) insertNewUpload(sessionId *string, filename string, vaultName *string) uint {
	upload := &domain.Upload{
		VaultName: *vaultName,
		Filename:  filename,
		SessionId: *sessionId,
		CreatedAt: util.GetDBNowStr(),
		Status:    domain.STARTED,
	}

	if err := s.dao.InsertUpload(upload); err != nil {
		fmt.Printf("Insert upload failed for %s and session %s.\n", *sessionId, filename)
	}
	return upload.ID
}

func (s *archiveUploadServiceImpl) updateCompletedUpload(id uint, res *domain.CompleteMultipartUploadOutput) {
	upload := s.dao.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("Failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Location = *res.Location
	upload.Checksum = *res.Checksum
	upload.ArchiveId = *res.ArchiveID
	upload.Status = domain.COMPLETED
	s.dao.UpdateUpload(upload)
}

func (s *archiveUploadServiceImpl) updateFailedUpload(id uint) {
	upload := s.dao.GetUploadByID(id)
	if upload == nil {
		fmt.Printf("failed to update upload %d: record not found.\n", id)
		return
	}

	upload.Status = domain.FAILED
	s.dao.UpdateUpload(upload)
}

func (s *archiveUploadServiceImpl) insertUploadedSegment(checksum *string, segNum int, segCount int, uploadId uint) {
	if uploadId == 0 {
		return
	}

	seg := &domain.UploadedSegment{
		SegmentNum: segNum,
		UploadId:   uploadId,
		Checksum:   *checksum,
		CreatedAt:  util.GetDBNowStr(),
	}

	if err := s.dao.InsertUploadedSegment(seg); err != nil {
		fmt.Printf("Insert uploaded segment failed for seg num %d, and upload id %d.\n", segNum, uploadId)
	}
}
