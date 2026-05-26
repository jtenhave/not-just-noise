package transactionaljob

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type transactionManagerMock struct {
	mock.Mock
}

func (m *transactionManagerMock) WithinTx(ctx context.Context, transaction func(context.Context) error) error {
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

type transactionalJobsRepoMock struct {
	mock.Mock
}

func (m *transactionalJobsRepoMock) GetAvailableTransactionalJobs(ctx context.Context, limit int) ([]TransactionalJob, error) {
	arguments := m.Called(ctx, limit)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	var jobs []TransactionalJob
	if j := arguments.Get(0); j != nil {
		jobs = j.([]TransactionalJob)
	}

	return jobs, err
}

func (m *transactionalJobsRepoMock) ClaimTransactionalJobs(ctx context.Context, ids []string) error {
	arguments := m.Called(ctx, ids)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *transactionalJobsRepoMock) ReleaseTransactionalJob(ctx context.Context, id string, lastError string) error {
	arguments := m.Called(ctx, id, lastError)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *transactionalJobsRepoMock) DeleteTransactionalJob(ctx context.Context, id string) error {
	arguments := m.Called(ctx, id)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

type transactionalJobHandlerMock struct {
	mock.Mock
}

func (m *transactionalJobHandlerMock) Handle(ctx context.Context, destination string, payload string) error {
	arguments := m.Called(ctx, destination, payload)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}

	return err
}

func TestTransactionalJobService_GetAvailableTransactionalJobs_Success_NoJobs(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	handlers := map[string]TransactionalJobHandler{
		"handler": new(transactionalJobHandlerMock),
	}

	service := NewTransactionalJobService(tm, repo, handlers)

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableTransactionalJobs", mock.Anything, 10).Return([]TransactionalJob(nil), nil)

	jobs, err := service.GetAvailableTransactionalJobs(context.Background(), 10)
	assert.NoError(t, err)
	assert.Nil(t, jobs)

	repo.AssertNotCalled(t, "ClaimTransactionalJobs")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_GetAvailableTransactionalJobs_Success_ClaimsJobs(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	handlers := map[string]TransactionalJobHandler{
		"handler": new(transactionalJobHandlerMock),
	}

	service := NewTransactionalJobService(tm, repo, handlers)

	expectedJobs := []TransactionalJob{
		{ID: "1", CallbackType: "handler", CallbackResource: "https://test.com/callback/1", Payload: "payload1"},
		{ID: "2", CallbackType: "handler", CallbackResource: "https://test.com/callback/2", Payload: "payload2"},
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableTransactionalJobs", mock.Anything, 2).Return(expectedJobs, nil)
	repo.On("ClaimTransactionalJobs", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ids := args.Get(1).([]string)
			assert.Equal(t, []string{"1", "2"}, ids)
		}).
		Return(nil)

	jobs, err := service.GetAvailableTransactionalJobs(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobs, jobs)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_GetAvailableTransactionalJobs_RepoError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	handlers := map[string]TransactionalJobHandler{
		"handler": new(transactionalJobHandlerMock),
	}

	service := NewTransactionalJobService(tm, repo, handlers)

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableTransactionalJobs", mock.Anything, 10).Return([]TransactionalJob(nil), errors.New("repo error"))

	_, err := service.GetAvailableTransactionalJobs(context.Background(), 10)
	assert.Error(t, err)

	repo.AssertNotCalled(t, "ClaimTransactionalJobs")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_GetAvailableTransactionalJobs_ClaimError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	handlers := map[string]TransactionalJobHandler{
		"handler": new(transactionalJobHandlerMock),
	}

	service := NewTransactionalJobService(tm, repo, handlers)

	expectedJobs := []TransactionalJob{
		{ID: "1", CallbackType: "handler", CallbackResource: "https://test.com/callback/1", Payload: "payload1"},
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableTransactionalJobs", mock.Anything, 10).Return(expectedJobs, nil)
	repo.On("ClaimTransactionalJobs", mock.Anything, []string{"1"}).Return(errors.New("claim error"))

	_, err := service.GetAvailableTransactionalJobs(context.Background(), 10)
	assert.Error(t, err)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_SendTransactionalJob_NoHandlerFound(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	handler := new(transactionalJobHandlerMock)
	handlers := map[string]TransactionalJobHandler{
		"handler": handler,
	}

	service := NewTransactionalJobService(tm, repo, handlers)

	job := TransactionalJob{
		ID:               "123",
		CallbackType:     "no-handler",
		CallbackResource: "https://test.com/callback",
		Payload:          "payload",
	}

	repo.On("ReleaseTransactionalJob", mock.Anything, job.ID, mock.Anything).Return(nil)

	err := service.SendTransactionalJob(context.Background(), job)
	assert.NoError(t, err)

	repo.AssertCalled(t, "ReleaseTransactionalJob", mock.Anything, job.ID, mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "no handler found for callback type: no-handler")
	}))
	repo.AssertNotCalled(t, "DeleteTransactionalJob")
	handler.AssertNotCalled(t, "Handle")
	handler.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_SendTransactionalJob_HandlerError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	handler := new(transactionalJobHandlerMock)
	handlers := map[string]TransactionalJobHandler{
		"handler": handler,
	}

	service := NewTransactionalJobService(tm, repo, handlers)

	job := TransactionalJob{
		ID:               "123",
		CallbackType:     "handler",
		CallbackResource: "https://test.com/callback",
		Payload:          "payload",
	}

	handler.On("Handle", mock.Anything, job.CallbackResource, job.Payload).Return(errors.New("handler error"))
	repo.On("ReleaseTransactionalJob", mock.Anything, job.ID, mock.Anything).Return(nil)

	err := service.SendTransactionalJob(context.Background(), job)
	assert.NoError(t, err)

	repo.AssertCalled(t, "ReleaseTransactionalJob", mock.Anything, job.ID, mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "handler error")
	}))
	repo.AssertNotCalled(t, "DeleteTransactionalJob")
	handler.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_SendTransactionalJob_Success_DeletesJob(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	handler := new(transactionalJobHandlerMock)
	handlers := map[string]TransactionalJobHandler{
		"handler": handler,
	}

	service := NewTransactionalJobService(tm, repo, handlers)

	job := TransactionalJob{
		ID:               "123",
		CallbackType:     "handler",
		CallbackResource: "https://test.com/callback",
		Payload:          "payload",
	}

	handler.On("Handle", mock.Anything, job.CallbackResource, job.Payload).Return(nil)
	repo.On("DeleteTransactionalJob", mock.Anything, job.ID).Return(nil)

	err := service.SendTransactionalJob(context.Background(), job)
	assert.NoError(t, err)

	repo.AssertNotCalled(t, "ReleaseTransactionalJob")
	handler.AssertExpectations(t)
	repo.AssertExpectations(t)
}
