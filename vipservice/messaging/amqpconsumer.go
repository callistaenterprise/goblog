package messaging

import (
        "github.com/streadway/amqp"
        "fmt"
)

// Defines our interface for connecting and consuming messages.
type IMessagingConsumer interface {
        ConnectToBroker(connectionString string)
        Subscribe(queueName string, handlerFunc func(amqp.Delivery)) error
        Close()
}

// Real implementation, encapsulates a pointer to an amqp.Connection
type MessagingConsumer struct {
        conn *amqp.Connection
}

func (m *MessagingConsumer) ConnectToBroker(connectionString string) {
        if connectionString == "" {
                panic("Cannot initialize connection to broker, connectionString not set. Have you initialized?")
        }

        var err error
        m.conn, err = amqp.Dial(fmt.Sprintf("%s/", connectionString))
        if err != nil {
                panic("Failed to connect to AMQP compatible broker at: " + connectionString)
        }
}

func (m *MessagingConsumer) Subscribe(queueName string, handlerFunc func(amqp.Delivery)) error {
        ch, err := m.conn.Channel()
        failOnError(err, "Failed to open a channel")
        defer ch.Close()

        q, err := ch.QueueDeclare(
                "vipQueue", // name
                false,   // durable
                false,   // delete when usused
                false,   // exclusive
                false,   // no-wait
                nil,     // arguments
        )
        failOnError(err, "Failed to declare a queue")

        msgs, err := ch.Consume(
                q.Name, // queue
                "",     // consumer
                true,   // auto-ack
                false,  // exclusive
                false,  // no-local
                false,  // no-wait
                nil,    // args
        )
        failOnError(err, "Failed to register a consumer")

        go func() {
                for d := range msgs {
                        // Invoke the OnMessage func we passed as parameter.
                        handlerFunc(d)
                }
        }()
        return err
}

func (m *MessagingConsumer) Close() {
        if m.conn != nil {
                m.conn.Close()
        }
}

func failOnError(err error, msg string) {
        if err != nil {
                fmt.Printf("%s: %s", msg, err)
                panic(fmt.Sprintf("%s: %s", msg, err))
        }
}