package messaging

import "github.com/stretchr/testify/mock"

type MockMessagingClient struct {
        mock.Mock
}

func (m *MockMessagingClient) ConnectToBroker(connectionString string) {

}

func (m *MockMessagingClient) SendMessage(body []byte, contentType string, exchangeName string, exchangeType string) error {
        args := m.Called(body, contentType, exchangeName, exchangeType)
        return args.Error(0)
}

func (m *MockMessagingClient) Close() {
        m.Called()
}
