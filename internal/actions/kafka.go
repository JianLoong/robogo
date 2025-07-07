package actions

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const defaultKafkaConsumeTimeout = 20 * time.Second // Default timeout for Kafka consumer if not specified

func KafkaAction(args []interface{}) (map[string]interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("kafka action requires at least one argument")
	}

	subcommand, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("kafka subcommand must be a string")
	}

	switch subcommand {
	case "publish":
		return kafkaPublish(args)
	case "consume":
		return kafkaConsume(args)
	default:
		return nil, fmt.Errorf("unknown kafka subcommand: %s", subcommand)
	}
}

func kafkaPublish(args []interface{}) (map[string]interface{}, error) {
	if len(args) < 4 {
		return nil, fmt.Errorf("kafka publish requires a broker, topic, and message")
	}
	brokersStr, ok1 := args[1].(string)
	topic, ok2 := args[2].(string)
	message, ok3 := args[3].(string)
	if !ok1 || !ok2 || !ok3 {
		return nil, fmt.Errorf("kafka publish broker, topic, and message must be strings")
	}
	brokers := strings.Split(brokersStr, ",")

	var settings map[string]interface{}
	if len(args) > 4 {
		var ok bool
		settings, ok = args[4].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("kafka publish settings must be a map")
		}
	}

	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireOne, // Default
	}

	if settings != nil {
		if acks, ok := settings["acks"].(string); ok {
			switch strings.ToLower(acks) {
			case "all":
				w.RequiredAcks = kafka.RequireAll
			case "none", "0":
				w.RequiredAcks = kafka.RequireNone
			case "one", "1":
				w.RequiredAcks = kafka.RequireOne
			}
		}
		if comp, ok := settings["compression"].(string); ok {
			switch strings.ToLower(comp) {
			case "gzip":
				w.Compression = kafka.Gzip
			case "snappy":
				w.Compression = kafka.Snappy
			case "lz4":
				w.Compression = kafka.Lz4
			case "zstd":
				w.Compression = kafka.Zstd
			}
		}
	}

	defer w.Close()

	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Value: []byte(message),
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to write messages: %w", err)
	}

	return map[string]interface{}{"status": "message published"}, nil
}

func kafkaConsume(args []interface{}) (map[string]interface{}, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("kafka consume requires a broker and topic")
	}
	brokersStr, ok1 := args[1].(string)
	topic, ok2 := args[2].(string)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("kafka consume broker and topic must be strings")
	}
	brokers := strings.Split(brokersStr, ",")

	var settings map[string]interface{}
	if len(args) > 3 {
		var ok bool
		settings, ok = args[3].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("kafka consume settings must be a map")
		}
	}

	readerConfig := kafka.ReaderConfig{
		Brokers:  brokers,
		Topic:    topic,
		MinBytes: 1,    // 1 Byte
		MaxBytes: 10e6, // 10MB
		MaxWait:  10 * time.Second,
	}

	if settings != nil {
		if groupID, ok := settings["groupID"].(string); ok {
			readerConfig.GroupID = groupID
		}
		if fromOffset, ok := settings["fromOffset"].(string); ok {
			if strings.ToLower(fromOffset) == "first" {
				readerConfig.StartOffset = kafka.FirstOffset
			} else if strings.ToLower(fromOffset) == "last" {
				readerConfig.StartOffset = kafka.LastOffset
			}
		}
		if partition, ok := settings["partition"].(string); ok {
			p, err := strconv.Atoi(partition)
			if err == nil {
				readerConfig.Partition = p
			}
		}
	}

	// If GroupID is not set, we must specify a partition. Default to 0.
	if readerConfig.GroupID == "" {
		if settings == nil || (settings != nil && settings["partition"] == nil) {
			readerConfig.Partition = 0
		}
	}

	r := kafka.NewReader(readerConfig)
	defer r.Close()

	// Use default timeout unless overridden in settings
	timeout := defaultKafkaConsumeTimeout
	if settings != nil {
		if t, ok := settings["timeout"]; ok {
			switch v := t.(type) {
			case int:
				timeout = time.Duration(v) * time.Second
			case int64:
				timeout = time.Duration(v) * time.Second
			case float64:
				timeout = time.Duration(v) * time.Second
			case string:
				// Try to parse as Go duration string (e.g., "10s", "1m")
				if d, err := time.ParseDuration(v); err == nil {
					timeout = d
				} else if n, err := strconv.Atoi(v); err == nil {
					timeout = time.Duration(n) * time.Second
				}
			}
		}
	}

	// Determine count (number of messages to consume)
	count := 1
	if settings != nil {
		if c, ok := settings["count"]; ok {
			switch v := c.(type) {
			case int:
				if v > 1 {
					count = v
				}
			case int64:
				if v > 1 {
					count = int(v)
				}
			case float64:
				if v > 1 {
					count = int(v)
				}
			case string:
				if n, err := strconv.Atoi(v); err == nil && n > 1 {
					count = n
				}
			}
		}
	}

	// Hybrid batch consumption: up to 'count' messages, but never longer than 'timeout' in total
	messages := []map[string]interface{}{}
	start := time.Now()
	for i := 0; i < count; i++ {
		remaining := timeout - time.Since(start)
		if remaining <= 0 {
			break
		}
		ctx, cancel := context.WithTimeout(context.Background(), remaining)
		m, err := r.ReadMessage(ctx)
		cancel()
		if err != nil {
			// Improved error handling
			if len(messages) == 0 {
				if errors.Is(err, context.DeadlineExceeded) {
					return map[string]interface{}{
						"error":   "timeout",
						"message": fmt.Sprintf("No message received within %v", timeout),
						"waited":  timeout.String(),
						"topic":   topic,
					}, nil
				}
				return map[string]interface{}{
					"error":   "kafka_error",
					"message": err.Error(),
					"topic":   topic,
				}, nil
			}
			break
		}
		messages = append(messages, map[string]interface{}{
			"message":   string(m.Value),
			"topic":     m.Topic,
			"partition": m.Partition,
			"offset":    m.Offset,
		})
	}

	if count == 1 {
		return messages[0], nil
	}
	return map[string]interface{}{"messages": messages}, nil
}
