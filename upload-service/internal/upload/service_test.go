package upload

import (
	"context"
	"io"

	"github.com/jtenhave/not-just-noise/contracts/dispatch"
	"github.com/stretchr/testify/mock"
)

type txManagerMock struct {
	mock.Mock
}

func (m *txManagerMock) WithinTx(ctx context.Context, transaction func(context.Context) error) error {
	arguments := m.Called(ctx, transaction)

	// Run the transaction callback to mimic real behavior.
	if transaction != nil {
		_ = transaction(ctx)
	}

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

type uploadRepoMock struct {
	mock.Mock
}

func (m *uploadRepoMock) GetUpload(ctx context.Context, audioID string) (Upload, error) {
	arguments := m.Called(ctx, audioID)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	var upload Upload
	if j := arguments.Get(0); j != nil {
		upload = j.(Upload)
	}

	return upload, err
}

func (m *uploadRepoMock) CreateUpload(ctx context.Context, upload Upload) error {
	arguments := m.Called(ctx, upload)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *uploadRepoMock) UpdateUpload(ctx context.Context, upload Upload) error {
	arguments := m.Called(ctx, upload)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *uploadRepoMock) DeleteUpload(ctx context.Context, id string, version int64) error {
	arguments := m.Called(ctx, id, version)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

type fileDownloaderMock struct {
	mock.Mock
}

func (m *fileDownloaderMock) DownloadFile(ctx context.Context, url string) (io.ReadCloser, error) {
	arguments := m.Called(ctx, url)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).(io.ReadCloser), err
}


type storageClientMock struct {
	mock.Mock
}

func (m *storageClientMock) UploadFile(ctx context.Context, key string, body io.Reader, metadata map[string]string) error {
	arguments := m.Called(ctx, key, body, metadata)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}


type dispatchClientMock struct {
	mock.Mock
}

func (m *dispatchClientMock) Dispatch(ctx context.Context, dispatch dispatch.Dispatch) error {
	arguments := m.Called(ctx, dispatch)
	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

