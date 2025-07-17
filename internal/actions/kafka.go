package actions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
	"github.com/segmentio/kafka-go"
)

// Kafka action - simplified implementation with immediate connection management
func kafkaAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 2 {
		return types.NewErrorResult("kafka action requires at least 2 arguments: operation, broker")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	broker := fmt.Sprintf("%v", args[1])

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch operation {
	case "publish":
		if len(args) < 4 {
			return types.NewErrorResult("kafka publish requires: operation, broker, topic, message")
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
			return types.NewErrorResult("failed to publish message: %v", err)
		}
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   map[string]interface{}{"status": "published"},
		}, nil

	case "consume":
		if len(args) < 3 {
			return types.NewErrorResult("kafka consume requires: operation, broker, topic")
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

		count := 1
		if c, ok := options["count"]; ok {
			switch v := c.(type) {
			case int:
				count = v
			case int64:
				count = int(v)
			case float64:
				count = int(v)
			case string:
				if parsed, err := strconv.Atoi(v); err == nil {
					count = parsed
				}
			}
			if count < 1 {
				count = 1
			}
		}

		var messages []string
		var lastPartition int
		var lastOffset int64
		for i := 0; i < count; i++ {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if i == 0 {
					return types.NewErrorResult("failed to consume message: %v", err)
				}
				break // return what we have so far
			}
			messages = append(messages, string(m.Value))
			lastPartition = m.Partition
			lastOffset = m.Offset
		}

		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data: map[string]interface{}{
				"messages":  messages,
				"count":     len(messages),
				"partition": lastPartition,
				"offset":    lastOffset,
			},
			Output: fmt.Sprintf("Consumed %d Kafka message(s): %v", len(messages), messages),
		}, nil

	default:
		return types.NewErrorResult("unknown kafka operation: %s", operation)
	}
}
