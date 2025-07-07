package actions

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var rabbitmqConnections = make(map[string]*amqp.Connection)

func RabbitMQAction(args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("rabbitmq action requires at least one argument")
	}

	subcommand := args[0]
	switch subcommand {
	case "connect":
		return rabbitmqConnect(args[1:])
	case "publish":
		return rabbitmqPublish(args[1:])
	case "consume":
		return rabbitmqConsume(args[1:])
	case "close":
		return rabbitmqClose(args[1:])
	default:
		return nil, fmt.Errorf("unknown rabbitmq subcommand: %s", subcommand)
	}
}

func rabbitmqConnect(args []string) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("rabbitmq connect requires a connection string and a connection name")
	}
	connStr := args[0]
	connName := args[1]

	conn, err := amqp.Dial(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	rabbitmqConnections[connName] = conn

	return map[string]interface{}{"status": "connected", "connection_name": connName}, nil
}

func rabbitmqPublish(args []string) (interface{}, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("rabbitmq publish requires a connection name, exchange, routing key, and message body")
	}
	connName := args[0]
	exchange := args[1]
	routingKey := args[2]
	body := args[3]

	conn, ok := rabbitmqConnections[connName]
	if !ok {
		return nil, fmt.Errorf("rabbitmq connection not found: %s", connName)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = ch.PublishWithContext(ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		return nil, fmt.Errorf("failed to publish a message: %w", err)
	}

	return map[string]interface{}{"status": "message published"}, nil
}

func rabbitmqConsume(args []string) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("rabbitmq consume requires a connection name and a queue name")
	}
	connName := args[0]
	queueName := args[1]

	conn, ok := rabbitmqConnections[connName]
	if !ok {
		return nil, fmt.Errorf("rabbitmq connection not found: %s", connName)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register a consumer: %w", err)
	}

	select {
	case d := <-msgs:
		return map[string]interface{}{"message": string(d.Body)}, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("timed out waiting for message from queue: %s", queueName)
	}
}

func rabbitmqClose(args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("rabbitmq close requires a connection name")
	}
	connName := args[0]

	conn, ok := rabbitmqConnections[connName]
	if !ok {
		return nil, fmt.Errorf("rabbitmq connection not found: %s", connName)
	}

	err := conn.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close RabbitMQ connection: %w", err)
	}
	delete(rabbitmqConnections, connName)

	return map[string]interface{}{"status": "disconnected", "connection_name": connName}, nil
}
