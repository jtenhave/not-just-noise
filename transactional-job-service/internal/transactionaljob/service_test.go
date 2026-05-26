package transactionaljob

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jtenhave/not-just-noise/lib/http"
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

type httpClientMock struct {
	mock.Mock
}

func (m *httpClientMock) Post(ctx context.Context, url string, body interface{}) (http.Response, error) {
	arguments := m.Called(ctx, url, body)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).(http.Response), err
}

func TestTransactionalJobService_GetAvailableTransactionalJobs_Success_NoJobs(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	client := new(httpClientMock)

	service := NewTransactionalJobService(tm, repo, client)

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
	client := new(httpClientMock)

	service := NewTransactionalJobService(tm, repo, client)

	expectedJobs := []TransactionalJob{
		{ID: "1", CallbackURL: "https://test.com/callback/1", Payload: "payload1"},
		{ID: "2", CallbackURL: "https://test.com/callback/2", Payload: "payload2"},
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableTransactionalJobs", mock.Anything, 2).Return(expectedJobs, nil)
	repo.On("ClaimTransactionalJobs", mock.Anything, []string{"1", "2"}).Return(nil)

	jobs, err := service.GetAvailableTransactionalJobs(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, expectedJobs, jobs)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_GetAvailableTransactionalJobs_RepoError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	client := new(httpClientMock)

	service := NewTransactionalJobService(tm, repo, client)

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableTransactionalJobs", mock.Anything, 10).Return([]TransactionalJob(nil), errors.New("db down"))

	_, err := service.GetAvailableTransactionalJobs(context.Background(), 10)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to get available transactional jobs"))

	repo.AssertNotCalled(t, "ClaimTransactionalJobs")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_GetAvailableTransactionalJobs_ClaimError(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	client := new(httpClientMock)

	service := NewTransactionalJobService(tm, repo, client)

	expectedJobs := []TransactionalJob{
		{ID: "1", CallbackURL: "https://test.com/callback/1", Payload: "payload1"},
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableTransactionalJobs", mock.Anything, 10).Return(expectedJobs, nil)
	repo.On("ClaimTransactionalJobs", mock.Anything, []string{"1"}).Return(errors.New("lock wait timeout"))

	_, err := service.GetAvailableTransactionalJobs(context.Background(), 10)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "failed to claim transactional jobs"))

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_SendTransactionalJob_Success_DeletesJob(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	client := new(httpClientMock)

	service := NewTransactionalJobService(tm, repo, client)

	job := TransactionalJob{
		ID:          "123",
		CallbackURL: "https://test.com/callback",
		Payload:     "payload",
	}

	client.On("Post", mock.Anything, job.CallbackURL, job.Payload).Return(http.Response{Code: http.Ok, Body: nil}, nil)
	repo.On("DeleteTransactionalJob", mock.Anything, job.ID).Return(nil)

	err := service.SendTransactionalJob(context.Background(), job)
	assert.NoError(t, err)

	repo.AssertNotCalled(t, "ReleaseTransactionalJob")
	client.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_SendTransactionalJob_CallbackError_ReleasesJob(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	client := new(httpClientMock)

	service := NewTransactionalJobService(tm, repo, client)

	job := TransactionalJob{
		ID:          "123",
		CallbackURL: "https://test.com/callback",
		Payload:     "payload",
	}

	body := `{"error":"test error message"}`
	client.On("Post", mock.Anything, job.CallbackURL, job.Payload).Return(http.Response{Code: 500, Body: &body}, nil)
	repo.On("ReleaseTransactionalJob", mock.Anything, job.ID, mock.Anything).Return(nil)

	err := service.SendTransactionalJob(context.Background(), job)
	assert.NoError(t, err)

	repo.AssertCalled(t, "ReleaseTransactionalJob", mock.Anything, job.ID, mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "callback returned code: 500") && strings.Contains(msg, "test error message")
	}))
	repo.AssertNotCalled(t, "DeleteTransactionalJob")
	client.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestTransactionalJobService_SendTransactionalJob_HTTPPostError_ReleasesJob(t *testing.T) {
	tm := new(transactionManagerMock)
	repo := new(transactionalJobsRepoMock)
	client := new(httpClientMock)

	service := NewTransactionalJobService(tm, repo, client)

	job := TransactionalJob{
		ID:          "123",
		CallbackURL: "https://test.com/callback",
		Payload:     "payload",
	}

	client.On("Post", mock.Anything, job.CallbackURL, job.Payload).Return(http.Response{}, errors.New("network error"))
	repo.On("ReleaseTransactionalJob", mock.Anything, job.ID, mock.Anything).Return(nil)

	err := service.SendTransactionalJob(context.Background(), job)
	assert.NoError(t, err)

	repo.AssertCalled(t, "ReleaseTransactionalJob", mock.Anything, job.ID, mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "callback returned code: 0") && strings.Contains(msg, "network error")
	}))
	repo.AssertNotCalled(t, "DeleteTransactionalJob")
	client.AssertExpectations(t)
	repo.AssertExpectations(t)
}