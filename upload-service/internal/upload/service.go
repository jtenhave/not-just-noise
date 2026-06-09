package upload

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io"
	"strconv"

	"github.com/google/uuid"
	"github.com/jtenhave/not-just-noise/contracts/dispatch"
	"github.com/jtenhave/not-just-noise/lib/log"
	"github.com/jtenhave/not-just-noise/lib/njnerror"
	uploadContract "github.com/jtenhave/not-just-noise/upload-service/internal/contracts"
)

type TxManager interface {
	WithinTx(ctx context.Context, transaction func(context.Context) error) error
}

type UploadRepo interface {
	GetUpload(ctx context.Context, audioID string) (Upload, error)
	CreateUpload(ctx context.Context, upload Upload) error
	UpdateUpload(ctx context.Context, upload Upload) error
	DeleteUpload(ctx context.Context, id string, version int64) error
}

type FileDownloader interface {
	DownloadFile(ctx context.Context, url string) (io.ReadCloser, int64, error)
}

type StorageClient interface {
	UploadFile(ctx context.Context, key string, body io.Reader, fileSize int64, metadata map[string]string) error
}

type DispatchClient interface {
	Dispatch(ctx context.Context, dispatch dispatch.Dispatch) error
}

type uploadService struct {
	transactionManager              TxManager
	uploadRepo                      UploadRepo
	fileDownloader                  FileDownloader
	tempStorageClient               StorageClient
	dispatchClient                  DispatchClient
	uploadStartedMessageDestination string
}

// NewUploadService creates a new uploadService using the given transactionManager and uploadRepo.
func NewUploadService(transactionManager TxManager, uploadRepo UploadRepo, fileDownloader FileDownloader, tempStorageClient StorageClient, dispatchClient DispatchClient, uploadStartedMessageDestination string) uploadService {
	return uploadService{
		transactionManager:              transactionManager,
		uploadRepo:                      uploadRepo,
		fileDownloader:                  fileDownloader,
		tempStorageClient:               tempStorageClient,
		dispatchClient:                  dispatchClient,
		uploadStartedMessageDestination: uploadStartedMessageDestination,
	}
}

func (uploadService uploadService) StartUpload(ctx context.Context, upload Upload) error {
	var existingUploadPtr *Upload
	existingUpload, err := uploadService.uploadRepo.GetUpload(ctx, upload.AudioID)
	if err != nil {
		if njnerror.Type(err) != njnerror.NotFound {
			return njnerror.Wrapf("uploadservice.StartUpload: failed to get upload record: %w", err)
		}
	} else {
		existingUploadPtr = &existingUpload
	}

	if existingUploadPtr != nil && existingUploadPtr.Version >= upload.Version {
		log.Logger(ctx).Info("upload version is outdated")
		return nil
	}

	fileReader, fileSize, err := uploadService.fileDownloader.DownloadFile(ctx, upload.FileURL)
	if err != nil {
		return njnerror.Wrapf("uploadservice.StartUpload: failed to download remote file: %w", err)
	}

	hashWriter := md5.New()
	teeReader := io.TeeReader(fileReader, hashWriter)

	metadata := map[string]string{
		"audio_id": upload.AudioID,
		"version":  strconv.FormatInt(upload.Version, 10),
	}

	tempFileKey := uuid.New().String() + ".mp3"
	err = uploadService.tempStorageClient.UploadFile(ctx, tempFileKey, teeReader, fileSize, metadata)
	if err != nil {
		return njnerror.Wrapf("uploadservice.StartUpload: failed to upload file: %w", err)
	}

	// Close explicitly so the hash writer is closed and the hash is calculated.
	fileReader.Close()

	fileHash := hex.EncodeToString(hashWriter.Sum(nil))
	if existingUploadPtr != nil && fileHash == existingUploadPtr.FileHash {
		log.Logger(ctx).Info("file hash is the same as the existing upload")
		return nil
	}

	upload.FileURL = tempFileKey
	upload.FileHash = fileHash

	err = uploadService.transactionManager.WithinTx(ctx, func(ctx context.Context) error {
		if existingUploadPtr == nil {
			// There may be a race condition here where the upload is already created. The incoming message will be retried and handled properly if that is the case.
			upload.ID = uuid.New().String()
			err = uploadService.uploadRepo.CreateUpload(ctx, upload)
			if err != nil {
				return njnerror.Wrapf("uploadservice.StartUpload: failed to create upload record: %w", err)
			}
		} else {
			upload.ID = existingUploadPtr.ID
			err = uploadService.uploadRepo.UpdateUpload(ctx, upload)
			if err != nil {
				return njnerror.Wrapf("uploadservice.StartUpload: failed to update upload record: %w", err)
			}
		}

		err = uploadService.sendUploadStartedMessage(ctx, upload)
		if err != nil {
			return njnerror.Wrapf("uploadservice.StartUpload: failed to send upload started message: %w", err)
		}

		return nil
	})

	if err != nil {
		// Return an error so that the
		return njnerror.Wrapf("uploadservice.StartUpload: failed to create upload record: %w", err)
	}

	log.Logger(ctx).Info("upload started", "upload", upload)

	return nil
}

// DeleteUpload deletes an upload record using the given upload. Returns the first error encountered.
func (uploadService uploadService) DeleteUpload(ctx context.Context, upload Upload) error {
	err := uploadService.uploadRepo.DeleteUpload(ctx, upload.ID, upload.Version)
	if err != nil {
		return njnerror.Wrapf("uploadservice.DeleteUpload: failed to delete upload record: %w", err)
	}

	return nil
}

// sendUploadStartedMessage sends a new upload started message for the given upload. Returns the first error encountered.
func (uploadService uploadService) sendUploadStartedMessage(ctx context.Context, upload Upload) error {
	uploadStartedMessage := uploadContract.UploadStartedMessage{
		UploadID: upload.ID,
		FileURL:  upload.FileURL,
		FileHash: upload.FileHash,
		Version:  upload.Version,
	}

	payloadJSON, err := json.Marshal(uploadStartedMessage)
	if err != nil {
		return njnerror.Wrapf("uploadservice.sendUploadStartedMessage: failed to create upload started message payload: %w", err)
	}

	dispatch := dispatch.Dispatch{
		ID:               uuid.New().String(),
		CallbackType:     dispatch.CallbackTypeLog, // TODO: Change to queue when the queue is implemented.
		CallbackResource: uploadService.uploadStartedMessageDestination,
		Payload:          string(payloadJSON),
	}

	err = uploadService.dispatchClient.Dispatch(ctx, dispatch)
	if err != nil {
		return njnerror.Wrapf("uploadservice.sendUploadStartedMessage: failed to dispatch upload started message: %w", err)
	}

	return nil
}
