package messaging

import (
        "github.com/streadway/amqp"
        "fmt"
        "log"
)

// Defines our interface for connecting and consuming messages.
type IMessagingConsumer interface {
        ConnectToBroker(connectionString string)
        Subscribe(exchangeName string, exchangeType string, consumerName string, handlerFunc func(amqp.Delivery)) error
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

func (m *MessagingConsumer) Subscribe(exchangeName string, exchangeType string, consumerName string, handlerFunc func(amqp.Delivery)) error {
        ch, err := m.conn.Channel()
        failOnError(err, "Failed to open a channel")
        defer ch.Close()

        err = ch.ExchangeDeclare(
                exchangeName,     // name of the exchange
                exchangeType, // type
                true,         // durable
                false,        // delete when complete
                false,        // internal
                false,        // noWait
                nil,          // arguments
        )
        failOnError(err, "Failed to register an Exchange")

        log.Printf("declared Exchange, declaring Queue (%s)", "")
        queue, err := ch.QueueDeclare(
                "", // name of the queue
                false,  // durable
                false, // delete when usused
                false, // exclusive
                false, // noWait
                nil,   // arguments
        )
        failOnError(err, "Failed to register an Queue")

        log.Printf("declared Queue (%d messages, %d consumers), binding to Exchange (key '%s')",
                queue.Messages, queue.Consumers, exchangeName)

        err = ch.QueueBind(
                queue.Name, // name of the queue
                exchangeName, // bindingKey
                exchangeName, // sourceExchange
                false, // noWait
                nil, // arguments
        );
        if err != nil {
                return fmt.Errorf("Queue Bind: %s", err)
        }

        msgs, err := ch.Consume(
                queue.Name, // queue
                consumerName, // consumer
                true,   // auto-ack
                false,  // exclusive
                false,  // no-local
                false,  // no-wait
                nil,    // args
        )
        failOnError(err, "Failed to register a consumer")

        go consumeLoop(msgs, handlerFunc)
        return nil
}

func consumeLoop(deliveries <-chan amqp.Delivery, handlerFunc func(d amqp.Delivery)) {
        for d := range deliveries {
                // Invoke the OnMessage func we passed as parameter.
                handlerFunc(d)
        }
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