package actions

import (
	"context"
	"encoding/json"
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
		return types.MissingArgsError("kafka", 2, len(args))
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	broker := fmt.Sprintf("%v", args[1])

	timeout := 30 * time.Second
	if timeoutOpt, ok := options["timeout"]; ok {
		switch t := timeoutOpt.(type) {
		case string:
			if duration, err := time.ParseDuration(t); err == nil {
				timeout = duration
			}
		case int, int64, float64:
			// Treat numbers as seconds
			var seconds float64
			switch num := t.(type) {
			case int:
				seconds = float64(num)
			case int64:
				seconds = float64(num)
			case float64:
				seconds = num
			}
			if seconds > 0 {
				timeout = time.Duration(seconds * float64(time.Second))
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	switch operation {
	case constants.OperationPublish:
		if len(args) < 4 {
			return types.MissingArgsError("kafka publish", 4, len(args))
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
			return types.RequestError(fmt.Sprintf("kafka publish to %s/%s", broker, topic), err.Error())
		}
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   map[string]any{"status": "published"},
		}

	case constants.OperationConsume:
		if len(args) < 3 {
			return types.MissingArgsError("kafka consume", 3, len(args))
		}
		topic := fmt.Sprintf("%v", args[2])

		config := kafka.ReaderConfig{
			Brokers:   []string{broker},
			Topic:     topic,
			Partition: 0,
			MinBytes:  1,
			MaxBytes:  10e6,
		}

		// Check for offset option
		if offsetOpt, ok := options["offset"]; ok {
			switch offset := offsetOpt.(type) {
			case string:
				switch strings.ToLower(offset) {
				case "earliest", "beginning":
					config.StartOffset = kafka.FirstOffset
				case "latest", "end":
					config.StartOffset = kafka.LastOffset
				default:
					// Try to parse as numeric offset
					if numOffset, err := strconv.ParseInt(offset, 10, 64); err == nil {
						config.StartOffset = numOffset
					}
				}
			case int:
				config.StartOffset = int64(offset)
			case int64:
				config.StartOffset = offset
			case float64:
				config.StartOffset = int64(offset)
			}
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
				return types.InvalidArgError("kafka consume", "count", "non-negative number")
			}
			if count == 0 {
				// Return empty result without consuming
				return types.ActionResult{
					Status: constants.ActionStatusPassed,
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
					// Check for topic-related errors first
					if strings.Contains(err.Error(), "UnknownTopicOrPartition") ||
					   strings.Contains(err.Error(), "topic does not exist") ||
					   strings.Contains(err.Error(), "UNKNOWN_TOPIC_OR_PARTITION") {
						return types.RequestError(fmt.Sprintf("kafka topic '%s' not found on broker %s", topic, broker), 
							"Topic may not exist or you may not have permission to access it")
					}
					
					// Check if it's a timeout error - could indicate topic doesn't exist
					if strings.Contains(err.Error(), "context deadline exceeded") {
						return types.TimeoutError(fmt.Sprintf("kafka consume from %s/%s timed out - check if topic exists and has messages", broker, topic))
					}
					
					// Check for authentication/authorization errors
					if strings.Contains(err.Error(), "SASL") || strings.Contains(err.Error(), "authentication") {
						return types.RequestError(fmt.Sprintf("kafka authentication failed for %s/%s", broker, topic), err.Error())
					}
					
					// Check for connection errors
					if strings.Contains(err.Error(), "connection refused") || 
					   strings.Contains(err.Error(), "no such host") {
						return types.RequestError(fmt.Sprintf("kafka broker %s unreachable", broker), 
							"Check if Kafka is running and broker address is correct")
					}
					
					return types.RequestError(fmt.Sprintf("kafka consume from %s/%s", broker, topic), err.Error())
				}
				break // return what we have so far
			}
			messages = append(messages, string(m.Value))
			lastPartition = m.Partition
			lastOffset = m.Offset
		}

		// Create the initial result structure
		resultData := map[string]any{
			"messages":  messages,
			"count":     len(messages),
			"partition": lastPartition,
			"offset":    lastOffset,
		}
		
		// Marshal and unmarshal to ensure JSON compatibility for jq
		jsonBytes, err := json.Marshal(resultData)
		if err != nil {
			return types.RequestError(fmt.Sprintf("kafka consume from %s/%s", broker, topic), fmt.Sprintf("JSON marshal error: %v", err))
		}
		
		var jsonCompatibleResult map[string]any
		if err := json.Unmarshal(jsonBytes, &jsonCompatibleResult); err != nil {
			return types.RequestError(fmt.Sprintf("kafka consume from %s/%s", broker, topic), fmt.Sprintf("JSON unmarshal error: %v", err))
		}

		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   jsonCompatibleResult,
		}

	default:
		return types.UnknownOperationError("kafka", operation)
	}
}
