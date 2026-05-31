package dispatch

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockDB struct {
	mock.Mock
}

func (m *mockDB) QueryContext(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	arguments := m.Called(ctx, query, args)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).([]map[string]any), err
}

func (m *mockDB) ExecContext(ctx context.Context, query string, args ...any) (int64, error) {
	arguments := m.Called(ctx, query, args)

	var err error
	if e := arguments.Get(1); e != nil {
		err = e.(error)
	}

	return arguments.Get(0).(int64), err
}

func (m *mockDB) IsTx(ctx context.Context) bool {
	arguments := m.Called(ctx)
	return arguments.Get(0).(bool)
}

func TestDispatchRepo_GetAvailableDispatches_NotInTx(t *testing.T) {
	db := new(mockDB)
	repo := NewDispatchRepo(db)

	db.On("IsTx", mock.Anything).Return(false)

	_, err := repo.GetAvailableDispatches(context.Background(), 10)

	assert.Error(t, err)
}

func TestDispatchRepo_GetAvailableDispatches_Success(t *testing.T) {
	db := new(mockDB)
	repo := NewDispatchRepo(db)

	expectedValues := []map[string]any{
		{
			"id":                "123",
			"callback_type":     "notify",
			"callback_resource": "https://test.com/callback",
			"payload":           "payload",
		},
	}

	db.On("IsTx", mock.Anything).Return(true)
	db.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return(expectedValues, nil)

	dispatches, err := repo.GetAvailableDispatches(context.Background(), 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(dispatches))
	assert.Equal(t, "notify", dispatches[0].CallbackType)
	assert.Equal(t, "https://test.com/callback", dispatches[0].CallbackResource)
	assert.Equal(t, "payload", dispatches[0].Payload)
}
