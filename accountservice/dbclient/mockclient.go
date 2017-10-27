package dbclient

import (
        "github.com/callistaenterprise/goblog/accountservice/model"
        "github.com/stretchr/testify/mock"
        "context"
)

// MockBoltClient is a mock implementation of a datastore client for testing purposes
type MockBoltClient struct {
        mock.Mock
}

// QueryAccount mock
func (m *MockBoltClient) QueryAccount(ctx context.Context, accountID string) (model.Account, error) {
        args := m.Mock.Called(ctx, accountID)
        return args.Get(0).(model.Account), args.Error(1)
}

// OpenBoltDb mock
func (m *MockBoltClient) OpenBoltDb() {
        // Does nothing
}

// Seed mock
func (m *MockBoltClient) Seed() {
        // Does nothing
}

// Check mock
func (m *MockBoltClient) Check() bool {
        args := m.Mock.Called()
        return args.Get(0).(bool)
}
