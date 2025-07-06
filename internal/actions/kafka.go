package actions

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

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
		if _, ok := settings["partition"]; !ok {
			readerConfig.Partition = 0
		}
	}

	r := kafka.NewReader(readerConfig)
	defer r.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	m, err := r.ReadMessage(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %w", err)
	}

	return map[string]interface{}{"message": string(m.Value), "topic": m.Topic, "partition": m.Partition, "offset": m.Offset}, nil
}
