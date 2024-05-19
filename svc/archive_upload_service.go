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
		upload, err := s.dao.GetUploadByID(ctx.UploadID)
		if err != nil {
			return err
		}

		uploadSessionId = &upload.SessionId
		id = upload.ID
	} else {
		var err error
		uploadSessionId, err = s.csp.InitiateMultipartUpload(ctx.ChunkSize, ctx.Vault)
		if err != nil {
			return err
		}
		fmt.Printf("Multipart upload session created, id: %s\n", *uploadSessionId)

		id, err = s.insertNewUpload(uploadSessionId, filename, ctx.Vault)
		if err != nil {
			return err
		}
	}

	finalChecksum, err := s.uploadSegments(uploadSessionId, id, filename, fl, ctx)
	if err != nil {
		return err
	}

	res, err := s.csp.CompleteMultipartUploadInput(fl, finalChecksum, uploadSessionId, ctx.Vault)
	if err != nil {
		return err
	}

	if err := s.updateCompletedUpload(id, res); err != nil {
		return err
	}

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
				if updateErr := s.updateFailedUpload(uploadId); updateErr != nil {
					fmt.Printf("unable to update failed upload: %s\n", updateErr)
				}

				if abortErr := s.csp.AbortMultipartUpload(uploadSessionId, ctx.Vault); abortErr != nil {
					fmt.Printf("unable to abort upload session: %s\n", abortErr)
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

			if err = s.insertUploadedSegment(&checksum, segNum, segCount, uploadId); err != nil {
				return nil, fmt.Errorf("insert uploaded segment failed for seg num %d, and upload id %d", segNum, uploadId)
			}
		}

		segNum += 1
		offset += int64(read)
	}

	encoded := toHexString(util.ComputeCombineHashChunks(hashes))
	return &encoded, nil
}

func (s *archiveUploadServiceImpl) insertNewUpload(
	sessionId *string,
	filename string,
	vaultName *string,
) (uint, error) {
	upload := &domain.Upload{
		VaultName: *vaultName,
		Filename:  filename,
		SessionId: *sessionId,
		CreatedAt: util.GetDBNowStr(),
		Status:    domain.STARTED,
	}

	if err := s.dao.InsertUpload(upload); err != nil {
		return 0, fmt.Errorf("insert upload %s failed for file %s", *sessionId, filename)
	}

	return upload.ID, nil
}

func (s *archiveUploadServiceImpl) updateCompletedUpload(id uint, res *domain.CompleteMultipartUploadOutput) error {
	upload, err := s.dao.GetUploadByID(id)
	if err != nil {
		return err
	}
	if upload == nil {
		return fmt.Errorf("failed to update upload %d: record not found", id)
	}

	upload.Location = *res.Location
	upload.Checksum = *res.Checksum
	upload.ArchiveId = *res.ArchiveID
	upload.Status = domain.COMPLETED
	return s.dao.UpdateUpload(upload)
}

func (s *archiveUploadServiceImpl) updateFailedUpload(id uint) error {
	upload, err := s.dao.GetUploadByID(id)
	if err != nil {
		return err
	}
	if upload == nil {
		return fmt.Errorf("failed to update upload %d: record not found", id)
	}

	upload.Status = domain.FAILED
	return s.dao.UpdateUpload(upload)
}

func (s *archiveUploadServiceImpl) insertUploadedSegment(
	checksum *string,
	segNum int,
	segCount int,
	uploadId uint,
) error {
	if uploadId == 0 {
		return nil
	}

	seg := &domain.UploadedSegment{
		SegmentNum: segNum,
		UploadId:   uploadId,
		Checksum:   *checksum,
		CreatedAt:  util.GetDBNowStr(),
	}

	return s.dao.InsertUploadedSegment(seg)
}
