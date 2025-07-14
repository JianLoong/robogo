package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/JianLoong/robogo/internal/common"
)

// No persistent connections - open, use, close immediately for CLI tool simplicity

// Kafka action - simplified implementation
func kafkaAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("kafka action requires at least 2 arguments: operation, broker")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	broker := fmt.Sprintf("%v", args[1])

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch operation {
	case "publish":
		if len(args) < 4 {
			return nil, fmt.Errorf("kafka publish requires: operation, broker, topic, message")
		}
		topic := fmt.Sprintf("%v", args[2])
		message := fmt.Sprintf("%v", args[3])

		w := &kafka.Writer{
			Addr:     kafka.TCP(broker),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
		defer w.Close()

		err := w.WriteMessages(ctx, kafka.Message{
			Value: []byte(message),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to publish message: %w", err)
		}
		return map[string]interface{}{"status": "published"}, nil

	case "consume":
		if len(args) < 3 {
			return nil, fmt.Errorf("kafka consume requires: operation, broker, topic")
		}
		topic := fmt.Sprintf("%v", args[2])

		r := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{broker},
			Topic:     topic,
			Partition: 0,
			MinBytes:  1,
			MaxBytes:  10e6,
		})
		defer r.Close()

		m, err := r.ReadMessage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to consume message: %w", err)
		}

		return map[string]interface{}{
			"message":   string(m.Value),
			"topic":     m.Topic,
			"partition": m.Partition,
			"offset":    m.Offset,
		}, nil

	default:
		return nil, fmt.Errorf("unknown kafka operation: %s", operation)
	}
}

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

// No cleanup needed - connections are closed after each operation