package actions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
	"github.com/segmentio/kafka-go"
)

// Kafka action - simplified implementation with immediate connection management
func kafkaAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 2 {
		return types.NewErrorResult("kafka action requires at least 2 arguments: operation, broker")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	broker := fmt.Sprintf("%v", args[1])

	timeout := 30 * time.Second
	if timeoutStr, ok := options["timeout"]; ok {
		if t, err := time.ParseDuration(fmt.Sprintf("%v", timeoutStr)); err == nil {
			timeout = t
		}
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch operation {
	case constants.OperationPublish:
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
			Data:   map[string]any{"status": "published"},
		}

	case constants.OperationConsume:
		if len(args) < 3 {
			return types.NewErrorResult("kafka consume requires: operation, broker, topic")
		}
		topic := fmt.Sprintf("%v", args[2])

		config := kafka.ReaderConfig{
			Brokers:   []string{broker},
			Topic:     topic,
			Partition: 0,
			MinBytes:  1,
			MaxBytes:  10e6,
		}

		// Check for auto-commit option
		if autoCommit, ok := options["auto_commit"]; ok {
			if enable, ok := autoCommit.(bool); ok && enable {
				config.GroupID = "robogo-consumer"
			}
		}

		r := kafka.NewReader(config)
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
			if count < 0 {
				return types.NewErrorResult("count must be >= 0, got %d", count)
			}
			if count == 0 {
				// Return empty result without consuming
				return types.ActionResult{
					Status: types.ActionStatusPassed,
					Data: map[string]any{
						"messages":  []string{},
						"count":     0,
						"partition": 0,
						"offset":    int64(0),
					},
				}
			}
		}

		var messages []string
		var lastPartition int
		var lastOffset int64
		for i := 0; i < count; i++ {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if i == 0 {
					// Check if it's a timeout error for better user feedback
					if strings.Contains(err.Error(), "context deadline exceeded") {
						return types.NewErrorResult("failed to consume message: timeout waiting for messages (topic may not exist or be empty)")
					}
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
			Data: map[string]any{
				"messages":  messages,
				"count":     len(messages),
				"partition": lastPartition,
				"offset":    lastOffset,
			},
		}

	default:
		return types.NewErrorResult("unknown kafka operation: %s", operation)
	}
}
