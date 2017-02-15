package messaging

import "github.com/stretchr/testify/mock"

type MockMessagingClient struct {
        mock.Mock
}

func (m *MockMessagingClient) ConnectToBroker(connectionString string) {

}

func (m *MockMessagingClient) SendMessage(body []byte, contentType string, queueName string) error {
        args := m.Called(body, contentType, queueName)
        return args.Error(0)
}

func (m *MockMessagingClient) Close() {
        m.Called()
}
