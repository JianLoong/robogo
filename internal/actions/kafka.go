package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/JianLoong/robogo/internal/common"
)

// Kafka action - simplified implementation with immediate connection management
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