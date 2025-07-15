package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/JianLoong/robogo/internal/common"
)

// RabbitMQ action - simplified implementation with proper resource management
func rabbitmqAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("rabbitmq action requires at least 3 arguments: operation, connection_string, queue/exchange")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])

	// Open connection for this operation only
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close() // Always close connection when done
	
	// Check connection health
	if conn.IsClosed() {
		return nil, fmt.Errorf("RabbitMQ connection closed immediately")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}
	defer func() {
		if closeErr := ch.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close RabbitMQ channel: %v\n", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch operation {
	case "publish":
		if len(args) < 5 {
			return nil, fmt.Errorf("rabbitmq publish requires: operation, connection, exchange, routing_key, message")
		}
		exchange := fmt.Sprintf("%v", args[2])
		routingKey := fmt.Sprintf("%v", args[3])
		message := fmt.Sprintf("%v", args[4])

		err = ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to publish message: %w", err)
		}
		return map[string]interface{}{"status": "published"}, nil

	case "consume":
		queueName := fmt.Sprintf("%v", args[2])
		msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to register consumer: %w", err)
		}

		select {
		case d := <-msgs:
			return map[string]interface{}{"message": string(d.Body)}, nil
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for message")
		}

	default:
		return nil, fmt.Errorf("unknown rabbitmq operation: %s", operation)
	}
}