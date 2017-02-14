package messaging

import (
        "fmt"
        "github.com/streadway/amqp"
)

// Defines our interface for connecting and sending messages.
type IMessagingClient interface {
        ConnectToBroker(connectionString string)
        SendMessage(msg []byte, contentType string, queueName string) error
}

// Real implementation, encapsulates a pointer to an amqp.Connection
type MessagingClient struct {
        conn *amqp.Connection
}

func (m *MessagingClient) ConnectToBroker(connectionString string) {
        if connectionString == "" {
                panic("Cannot initialize connection to broker, connectionString not set. Have you initialized?")
        }

        var err error
        m.conn, err = amqp.Dial(fmt.Sprintf("%s/", connectionString))
        if err != nil {
                panic("Failed to connect to AMQP compatible broker at: " + connectionString)
        }
}

func (m *MessagingClient) SendMessage(body []byte, contentType string, queueName string) error {
        if m.conn == nil {
                panic("Tried to send message before connection was initialized. Don't do that.")
        }
        ch, err := m.conn.Channel()      // Get a channel from the connection
        q, err := ch.QueueDeclare(       // Declare a queue that will be created if not exists with some args
                queueName, // our queue name
                false, // durable
                false, // delete when unused
                false, // exclusive
                false, // no-wait
                nil, // arguments
        )
        err = ch.Publish(                      // Publishes a message onto the queue.
                "", // exchange
                q.Name, // routing key
                false, // mandatory
                false, // immediate
                amqp.Publishing{
                        ContentType: contentType,
                        Body:        body,          // Our JSON body as []byte
                })
        return err
}
