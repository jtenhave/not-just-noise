package audioing

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jtenhave/not-just-noise/audio-service/internal/audio"
	audioNotify "github.com/jtenhave/not-just-noise/audio-service/internal/infrastructure/notify"
	"github.com/jtenhave/not-just-noise/audio-service/internal/infrastructure/repo"
	"github.com/jtenhave/not-just-noise/lib/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type audioRepoMock struct {
	mock.Mock
}

func (m *audioRepoMock) GetAudio(ctx context.Context, id string) (audio.Audio, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(audio.Audio), args.Error(1)
}

func (m *audioRepoMock) CreateAudio(ctx context.Context, a audio.Audio) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m *audioRepoMock) UpdateAudio(ctx context.Context, a audio.UpdateAudio, version int64) error {
	args := m.Called(ctx, a, version)
	return args.Error(0)
}

func (m *audioRepoMock) DeleteAudio(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type transactionalOutboxRepoMock struct {
	mock.Mock
}

func (m *transactionalOutboxRepoMock) CreateTransactionalOutboxRecord(ctx context.Context, record notify.TransactionalOutboxRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

type connectionRepositoryMock struct {
	mock.Mock
	audioRepo *audioRepoMock
}

func (m *connectionRepositoryMock) AudioRepo() repo.AudioRepo {
	return m.audioRepo
}

func (m *connectionRepositoryMock) TransactionalOutboxRepo() repo.TransactionalOutboxRepo {
	panic("connection transactional outbox repo should not be used in audio service tests")
}

func (m *connectionRepositoryMock) BeginTx(ctx context.Context) (repo.TransactionRepository, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repo.TransactionRepository), args.Error(1)
}

type transactionRepositoryMock struct {
	mock.Mock
	audioRepo  *audioRepoMock
	outboxRepo *transactionalOutboxRepoMock
}

func (m *transactionRepositoryMock) AudioRepo() repo.AudioRepo {
	return m.audioRepo
}

func (m *transactionRepositoryMock) TransactionalOutboxRepo() repo.TransactionalOutboxRepo {
	return m.outboxRepo
}

func (m *transactionRepositoryMock) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *transactionRepositoryMock) Rollback() error {
	args := m.Called()
	return args.Error(0)
}

type audioNotifyFormatterMock struct {
	mock.Mock
}

func (m *audioNotifyFormatterMock) NotifyAudioChangedPayload(eventType audio.AudioNotifierEventType, audio audio.Audio) (string, error) {
	args := m.Called(eventType, audio)
	return args.Get(0).(string), args.Error(1)
}

type testFixtures struct {
	service             audioService
	connectionRepo      *connectionRepositoryMock
	connectionAudioRepo *audioRepoMock
	tx                  *transactionRepositoryMock
	txAudioRepo         *audioRepoMock
	txOutboxRepo        *transactionalOutboxRepoMock
	audioNotifyFormatter *audioNotifyFormatterMock
}

func newTestFixtures() *testFixtures {
	connectionAudioRepo := new(audioRepoMock)
	txAudioRepo := new(audioRepoMock)
	txOutboxRepo := new(transactionalOutboxRepoMock)
	tx := &transactionRepositoryMock{
		audioRepo:  txAudioRepo,
		outboxRepo: txOutboxRepo,
	}
	connectionRepo := &connectionRepositoryMock{
		audioRepo: connectionAudioRepo,
	}
	audioNotifyFormatter := new(audioNotifyFormatterMock)

	return &testFixtures{
		service:             NewAudioService(connectionRepo, audioNotifyFormatter),
		connectionRepo:      connectionRepo,
		connectionAudioRepo: connectionAudioRepo,
		tx:                  tx,
		txAudioRepo:         txAudioRepo,
		txOutboxRepo:        txOutboxRepo,
		audioNotifyFormatter: audioNotifyFormatter,
	}
}

func (f *testFixtures) expectTx() {
	f.connectionRepo.On("BeginTx", mock.Anything).Return(f.tx, nil)
	f.tx.On("Rollback").Return(nil)
}

func (f *testFixtures) expectTxCommitSuccess() {
	f.expectTx()
	f.tx.On("Commit").Return(nil)
}

func sampleStoredAudio() audio.Audio {
	return audio.Audio{
		ID:        "123abc",
		CreatorID: "456def",
		Title:     "Test Audio",
		FileURL:   "https://test.com/audio.mp3",
		Version:   3,
		Status:    "active",
	}
}

func TestGetAudio_Success(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	expected := sampleStoredAudio()

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(expected, nil)

	got, err := f.service.GetAudio(ctx, "123abc")

	assert.NoError(t, err)
	assert.Equal(t, expected, got)
	f.connectionAudioRepo.AssertExpectations(t)
}

func TestGetAudio_Failure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(audio.Audio{}, fmt.Errorf("failed to get audio"))

	_, err := f.service.GetAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.GetAudio: failed to get audio")
	f.connectionAudioRepo.AssertExpectations(t)
}

func TestCreateAudio_Success(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	input := audio.Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"}

	f.expectTxCommitSuccess()
	f.txAudioRepo.On("CreateAudio", ctx, mock.MatchedBy(func(a audio.Audio) bool {
		return a.CreatorID == input.CreatorID &&
			a.Title == input.Title &&
			a.FileURL == input.FileURL &&
			a.ID != ""
	})).Return(nil)
	var createdAudioID string
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.MatchedBy(func(record notify.TransactionalOutboxRecord) bool {
		if record.ID == "" || record.Payload == "" {
			return false
		}
		var message struct {
			EventType audio.AudioNotifierEventType `json:"event_type"`
			AudioID   string                       `json:"id"`
			Version   int64                        `json:"version"`
		}
		if err := json.Unmarshal([]byte(record.Payload), &message); err != nil {
			return false
		}
		createdAudioID = message.AudioID
		return message.EventType == audio.AudioCreatedEvent && message.AudioID != "" && message.Version == 0
	})).Return(nil)

	id, err := f.service.CreateAudio(ctx, input)

	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	assert.Equal(t, id, createdAudioID)
	f.connectionRepo.AssertExpectations(t)
	f.tx.AssertExpectations(t)
	f.txAudioRepo.AssertExpectations(t)
	f.txOutboxRepo.AssertExpectations(t)
}

func TestCreateAudio_BeginTxFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.connectionRepo.On("BeginTx", ctx).Return(nil, fmt.Errorf("failed to begin transaction"))

	_, err := f.service.CreateAudio(ctx, audio.Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.CreateAudio: failed to begin transaction")
	f.txAudioRepo.AssertNotCalled(t, "CreateAudio")
}

func TestCreateAudio_CreateAudioFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.expectTx()
	f.txAudioRepo.On("CreateAudio", ctx, mock.Anything).Return(fmt.Errorf("failed to create audio"))

	_, err := f.service.CreateAudio(ctx, audio.Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.CreateAudio: failed to create audio")
	f.tx.AssertExpectations(t)
	f.txOutboxRepo.AssertNotCalled(t, "CreateTransactionalOutboxRecord")
	f.tx.AssertNotCalled(t, "Commit")
}

func TestCreateAudio_OutboxFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.expectTx()
	f.txAudioRepo.On("CreateAudio", ctx, mock.Anything).Return(nil)
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.Anything).Return(fmt.Errorf("failed to create outbox record"))

	_, err := f.service.CreateAudio(ctx, audio.Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.CreateAudio: failed to create transactional outbox record")
	f.tx.AssertNotCalled(t, "Commit")
}

func TestCreateAudio_CommitFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.expectTx()
	f.txAudioRepo.On("CreateAudio", ctx, mock.Anything).Return(nil)
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.Anything).Return(nil)
	f.tx.On("Commit").Return(fmt.Errorf("failed to commit"))

	_, err := f.service.CreateAudio(ctx, audio.Audio{CreatorID: "456def", Title: "Test Audio", FileURL: "https://test.com/audio.mp3"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.CreateAudio: failed to commit transaction")
}

func TestUpdateAudio_Success(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	stored := sampleStoredAudio()
	title := "New Title"
	update := audio.UpdateAudio{ID: "123abc", Title: &title}

	expectedPayload, err := audioNotify.NotifyAudioChangedPayload(audio.AudioUpdatedEvent, audio.Audio{
		ID:        stored.ID,
		CreatorID: stored.CreatorID,
		Title:     title,
		FileURL:   stored.FileURL,
		Version:   stored.Version + 1,
		Status:    stored.Status,
	})
	assert.NoError(t, err)

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(stored, nil)
	f.expectTxCommitSuccess()
	f.txAudioRepo.On("UpdateAudio", ctx, update, stored.Version).Return(nil)
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.MatchedBy(func(record notify.TransactionalOutboxRecord) bool {
		return record.ID != "" && record.Payload == expectedPayload
	})).Return(nil)

	err = f.service.UpdateAudio(ctx, update)

	assert.NoError(t, err)
	f.connectionAudioRepo.AssertExpectations(t)
	f.txAudioRepo.AssertExpectations(t)
	f.txOutboxRepo.AssertExpectations(t)
}

func TestUpdateAudio_GetAudioFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	title := "New Title"

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(audio.Audio{}, fmt.Errorf("failed to get audio"))

	err := f.service.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audiorepo.UpdateAudio: failed to get audio")
	f.connectionRepo.AssertNotCalled(t, "BeginTx")
}

func TestUpdateAudio_BeginTxFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	title := "New Title"

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(sampleStoredAudio(), nil)
	f.connectionRepo.On("BeginTx", ctx).Return(nil, fmt.Errorf("failed to begin transaction"))

	err := f.service.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.CreateAudio: failed to begin transaction")
}

func TestUpdateAudio_UpdateAudioFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	title := "New Title"

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(sampleStoredAudio(), nil)
	f.expectTx()
	f.txAudioRepo.On("UpdateAudio", ctx, mock.Anything, int64(3)).Return(fmt.Errorf("failed to update audio"))

	err := f.service.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.CreateAudio: failed to create audio")
	f.tx.AssertNotCalled(t, "Commit")
}

func TestUpdateAudio_OutboxFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	title := "New Title"

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(sampleStoredAudio(), nil)
	f.expectTx()
	f.txAudioRepo.On("UpdateAudio", ctx, mock.Anything, int64(3)).Return(nil)
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.Anything).Return(fmt.Errorf("failed to create outbox record"))

	err := f.service.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.CreateAudio: failed to create transactional outbox record")
	f.tx.AssertNotCalled(t, "Commit")
}

func TestUpdateAudio_CommitFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	title := "New Title"

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(sampleStoredAudio(), nil)
	f.expectTx()
	f.txAudioRepo.On("UpdateAudio", ctx, mock.Anything, int64(3)).Return(nil)
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.Anything).Return(nil)
	f.tx.On("Commit").Return(fmt.Errorf("failed to commit"))

	err := f.service.UpdateAudio(ctx, audio.UpdateAudio{ID: "123abc", Title: &title})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.CreateAudio: failed to commit transaction")
}

func TestDeleteAudio_Success(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()
	stored := sampleStoredAudio()

	expectedPayload, err := audioNotify.NotifyAudioChangedPayload(audio.AudioDeletedEvent, audio.Audio{
		ID:        stored.ID,
		CreatorID: stored.CreatorID,
		Title:     stored.Title,
		FileURL:   stored.FileURL,
		Version:   stored.Version + 1,
		Status:    "deleted",
	})
	assert.NoError(t, err)

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(stored, nil)
	f.expectTxCommitSuccess()
	f.txAudioRepo.On("DeleteAudio", ctx, "123abc").Return(nil)
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.MatchedBy(func(record notify.TransactionalOutboxRecord) bool {
		return record.ID != "" && record.Payload == expectedPayload
	})).Return(nil)

	err = f.service.DeleteAudio(ctx, "123abc")

	assert.NoError(t, err)
	f.connectionAudioRepo.AssertExpectations(t)
	f.txAudioRepo.AssertExpectations(t)
	f.txOutboxRepo.AssertExpectations(t)
}

func TestDeleteAudio_GetAudioFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(audio.Audio{}, fmt.Errorf("failed to get audio"))

	err := f.service.DeleteAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audiorepo.DeleteAudio: failed to get audio")
	f.connectionRepo.AssertNotCalled(t, "BeginTx")
}

func TestDeleteAudio_BeginTxFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(sampleStoredAudio(), nil)
	f.connectionRepo.On("BeginTx", ctx).Return(nil, fmt.Errorf("failed to begin transaction"))

	err := f.service.DeleteAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.DeleteAudio: failed to begin transaction")
}

func TestDeleteAudio_DeleteAudioFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(sampleStoredAudio(), nil)
	f.expectTx()
	f.txAudioRepo.On("DeleteAudio", ctx, "123abc").Return(fmt.Errorf("failed to delete audio"))

	err := f.service.DeleteAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.DeleteAudio: failed to delete audio")
	f.tx.AssertNotCalled(t, "Commit")
}

func TestDeleteAudio_OutboxFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(sampleStoredAudio(), nil)
	f.expectTx()
	f.txAudioRepo.On("DeleteAudio", ctx, "123abc").Return(nil)
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.Anything).Return(fmt.Errorf("failed to create outbox record"))

	err := f.service.DeleteAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.DeleteAudio: failed to create transactional outbox record")
	f.tx.AssertNotCalled(t, "Commit")
}

func TestDeleteAudio_CommitFailure(t *testing.T) {
	f := newTestFixtures()
	ctx := context.Background()

	f.connectionAudioRepo.On("GetAudio", ctx, "123abc").Return(sampleStoredAudio(), nil)
	f.expectTx()
	f.txAudioRepo.On("DeleteAudio", ctx, "123abc").Return(nil)
	f.txOutboxRepo.On("CreateTransactionalOutboxRecord", ctx, mock.Anything).Return(nil)
	f.tx.On("Commit").Return(fmt.Errorf("failed to commit"))

	err := f.service.DeleteAudio(ctx, "123abc")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audioservice.DeleteAudio: failed to commit transaction")
}
