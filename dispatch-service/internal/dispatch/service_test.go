package dispatch

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

type dispatchRepoMock struct {
	mock.Mock
}

func (m *dispatchRepoMock) GetAvailableDispatches(ctx context.Context, limit int) ([]Dispatch, error) {
	arguments := m.Called(ctx, limit)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	var jobs []Dispatch
	if j := arguments.Get(0); j != nil {
		jobs = j.([]Dispatch)
	}

	return jobs, err
}

func (m *dispatchRepoMock) ClaimDispatches(ctx context.Context, ids []string) error {
	arguments := m.Called(ctx, ids)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *dispatchRepoMock) ReleaseDispatch(ctx context.Context, id string, lastError string) error {
	arguments := m.Called(ctx, id, lastError)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

func (m *dispatchRepoMock) DeleteDispatch(ctx context.Context, id string) error {
	arguments := m.Called(ctx, id)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}
	return err
}

type dispatcherMock struct {
	mock.Mock
}

func (m *dispatcherMock) Dispatch(ctx context.Context, dispatcherType string, destination string, payload string) error {
	arguments := m.Called(ctx, dispatcherType, destination, payload)

	var err error
	if e := arguments.Get(0); e != nil {
		err = e.(error)
	}

	return err
}

func TestDispatchService_GetAvailableDispatches_Success_NoDispatches(t *testing.T) {
	tm := new(txManagerMock)
	repo := new(dispatchRepoMock)
	dispatcher := new(dispatcherMock)

	service := NewDispatchService(tm, repo, dispatcher)

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableDispatches", mock.Anything, 10).Return([]Dispatch(nil), nil)

	dispatches, err := service.GetAvailableDispatches(context.Background(), 10)
	assert.NoError(t, err)
	assert.Nil(t, dispatches)

	repo.AssertNotCalled(t, "ClaimDispatches")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDispatchService_GetAvailableDispatches_Success_ClaimsDispatches(t *testing.T) {
	tm := new(txManagerMock)
	repo := new(dispatchRepoMock)
	dispatcher := new(dispatcherMock)

	service := NewDispatchService(tm, repo, dispatcher)

	expectedDispatches := []Dispatch{
		{ID: "1", CallbackType: "handler", CallbackResource: "https://test.com/callback/1", Payload: "payload1"},
		{ID: "2", CallbackType: "handler", CallbackResource: "https://test.com/callback/2", Payload: "payload2"},
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableDispatches", mock.Anything, 2).Return(expectedDispatches, nil)
	repo.On("ClaimDispatches", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ids := args.Get(1).([]string)
			assert.Equal(t, []string{"1", "2"}, ids)
		}).
		Return(nil)

	dispatches, err := service.GetAvailableDispatches(context.Background(), 2)
	assert.NoError(t, err)
	assert.Equal(t, expectedDispatches, dispatches)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDispatchService_GetAvailableDispatches_RepoError(t *testing.T) {
	tm := new(txManagerMock)
	repo := new(dispatchRepoMock)
	dispatcher := new(dispatcherMock)

	service := NewDispatchService(tm, repo, dispatcher)

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableDispatches", mock.Anything, 10).Return([]Dispatch(nil), errors.New("repo error"))

	_, err := service.GetAvailableDispatches(context.Background(), 10)
	assert.Error(t, err)

	repo.AssertNotCalled(t, "ClaimDispatches")
	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDispatchService_GetAvailableDispatches_ClaimError(t *testing.T) {
	tm := new(txManagerMock)
	repo := new(dispatchRepoMock)
	dispatcher := new(dispatcherMock)

	service := NewDispatchService(tm, repo, dispatcher)

	expectedDispatches := []Dispatch{
		{ID: "1", CallbackType: "handler", CallbackResource: "https://test.com/callback/1", Payload: "payload1"},
	}

	tm.On("WithinTx", mock.Anything, mock.Anything).Return(nil)
	repo.On("GetAvailableDispatches", mock.Anything, 10).Return(expectedDispatches, nil)
	repo.On("ClaimDispatches", mock.Anything, []string{"1"}).Return(errors.New("claim error"))

	_, err := service.GetAvailableDispatches(context.Background(), 10)
	assert.Error(t, err)

	tm.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDispatchService_Dispatch_DispatcherError(t *testing.T) {
	tm := new(txManagerMock)
	repo := new(dispatchRepoMock)
	dispatcher := new(dispatcherMock)

	service := NewDispatchService(tm, repo, dispatcher)

	dispatch := Dispatch{
		ID:               "123",
		CallbackType:     "handler",
		CallbackResource: "https://test.com/callback",
		Payload:          "payload",
	}

	dispatcher.On("Dispatch", mock.Anything, dispatch.CallbackType, dispatch.CallbackResource, dispatch.Payload).Return(errors.New("dispatcher error"))
	repo.On("ReleaseDispatch", mock.Anything, dispatch.ID, mock.Anything).Return(nil)

	err := service.Dispatch(context.Background(), dispatch)
	assert.NoError(t, err)

	repo.AssertCalled(t, "ReleaseDispatch", mock.Anything, dispatch.ID, mock.MatchedBy(func(msg string) bool {
		return strings.Contains(msg, "dispatcher error")
	}))
	repo.AssertNotCalled(t, "DeleteDispatch")
	dispatcher.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestDispatchService_Dispatch_Success_DeletesDispatch(t *testing.T) {
	tm := new(txManagerMock)
	repo := new(dispatchRepoMock)
	dispatcher := new(dispatcherMock)

	service := NewDispatchService(tm, repo, dispatcher)

	dispatch := Dispatch{
		ID:               "123",
		CallbackType:     "handler",
		CallbackResource: "https://test.com/callback",
		Payload:          "payload",
	}

	dispatcher.On("Dispatch", mock.Anything, dispatch.CallbackType, dispatch.CallbackResource, dispatch.Payload).Return(nil)
	repo.On("DeleteDispatch", mock.Anything, dispatch.ID).Return(nil)

	err := service.Dispatch(context.Background(), dispatch)
	assert.NoError(t, err)

	repo.AssertNotCalled(t, "ReleaseDispatch")
	dispatcher.AssertExpectations(t)
	repo.AssertExpectations(t)
}
