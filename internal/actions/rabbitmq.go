package actions

import (
	"context"
	"fmt"
	"time"

	"github.com/JianLoong/robogo/internal/util"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQAction performs RabbitMQ operations using ActionContext manager
func RabbitMQAction(ctx context.Context, args []interface{}, options map[string]interface{}, silent bool) (interface{}, error) {
	// Convert interface{} args to string args for compatibility
	strArgs := make([]string, len(args))
	for i, v := range args {
		strArgs[i] = fmt.Sprintf("%v", v)
	}
	
	return rabbitmqActionInternal(ctx, strArgs)
}

func rabbitmqActionInternal(ctx context.Context, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, util.NewValidationError("rabbitmq action requires at least one argument", map[string]interface{}{
			"args_count": len(args),
			"required":   1,
		}).WithAction("rabbitmq")
	}

	// Get RabbitMQ manager from context
	actionCtx := GetActionContext(ctx)
	if actionCtx.RabbitMQManager == nil {
		return nil, util.NewConfigurationError("RabbitMQ manager not initialized", "rabbit_manager", nil).WithAction("rabbitmq")
	}

	subcommand := args[0]

	// Set timeout from settings or default (30s)
	timeout := 30 * time.Second
	if len(args) > 2 {
		// If the last argument is a timeout string, use it
		if t, err := time.ParseDuration(args[len(args)-1]); err == nil {
			timeout = t
			args = args[:len(args)-1] // Remove timeout from args
		}
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch subcommand {
	case "connect":
		return rabbitmqConnect(ctx, actionCtx.RabbitMQManager, args[1:])
	case "publish":
		return rabbitmqPublish(ctx, actionCtx.RabbitMQManager, args[1:])
	case "consume":
		return rabbitmqConsume(ctx, actionCtx.RabbitMQManager, args[1:], timeout)
	case "close":
		return rabbitmqClose(ctx, actionCtx.RabbitMQManager, args[1:])
	default:
		return nil, util.NewValidationError("unknown rabbitmq subcommand", map[string]interface{}{
			"subcommand":        subcommand,
			"valid_subcommands": []string{"connect", "publish", "consume", "close"},
		}).WithAction("rabbitmq")
	}
}

func rabbitmqConnect(ctx context.Context, manager *RabbitMQManager, args []string) (interface{}, error) {
	if len(args) < 2 {
		return nil, util.NewValidationError("rabbitmq connect requires a connection string and a connection name", map[string]interface{}{
			"args_count": len(args),
			"required":   2,
		}).WithAction("rabbitmq").WithStep("connect")
	}
	connStr := args[0]
	connName := args[1]

	err := manager.Connect(connName, connStr)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"status": "connected", "connection_name": connName}, nil
}

func rabbitmqPublish(ctx context.Context, manager *RabbitMQManager, args []string) (interface{}, error) {
	if len(args) < 4 {
		return nil, util.NewValidationError("rabbitmq publish requires a connection name, exchange, routing key, and message body", map[string]interface{}{
			"args_count": len(args),
			"required":   4,
		}).WithAction("rabbitmq").WithStep("publish")
	}
	connName := args[0]
	exchange := args[1]
	routingKey := args[2]
	body := args[3]

	conn, ok := manager.GetConnection(connName)
	if !ok {
		return nil, util.NewValidationError("rabbitmq connection not found", map[string]interface{}{
			"connection_name":        connName,
			"available_connections": manager.ListConnections(),
		}).WithAction("rabbitmq").WithStep("publish")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, util.NewMessagingError("failed to open a channel", err, "rabbitmq").WithStep("publish").WithDetails(map[string]interface{}{
			"connection_name": connName,
			"exchange":        exchange,
			"routing_key":     routingKey,
		})
	}
	defer ch.Close()

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
		return nil, util.NewMessagingError("failed to publish a message", err, "rabbitmq").WithStep("publish").WithDetails(map[string]interface{}{
			"connection_name": connName,
			"exchange":        exchange,
			"routing_key":     routingKey,
			"message_size":    len(body),
		})
	}

	return map[string]interface{}{"status": "message published"}, nil
}

func rabbitmqConsume(ctx context.Context, manager *RabbitMQManager, args []string, timeout time.Duration) (interface{}, error) {
	if len(args) < 2 {
		return nil, util.NewValidationError("rabbitmq consume requires a connection name and a queue name", map[string]interface{}{
			"args_count": len(args),
			"required":   2,
		}).WithAction("rabbitmq").WithStep("consume")
	}
	connName := args[0]
	queueName := args[1]

	conn, ok := manager.GetConnection(connName)
	if !ok {
		return nil, util.NewValidationError("rabbitmq connection not found", map[string]interface{}{
			"connection_name":        connName,
			"available_connections": manager.ListConnections(),
		}).WithAction("rabbitmq").WithStep("consume")
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, util.NewMessagingError("failed to open a channel", err, "rabbitmq").WithStep("consume").WithDetails(map[string]interface{}{
			"connection_name": connName,
			"queue_name":      queueName,
		})
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
		return nil, util.NewMessagingError("failed to register a consumer", err, "rabbitmq").WithStep("consume").WithDetails(map[string]interface{}{
			"connection_name": connName,
			"queue_name":      queueName,
		})
	}

	select {
	case d := <-msgs:
		return map[string]interface{}{"message": string(d.Body)}, nil
	case <-ctx.Done():
		return nil, util.NewTimeoutError("timed out waiting for message from queue", nil, "rabbitmq").WithStep("consume").WithDetails(map[string]interface{}{
			"connection_name": connName,
			"queue_name":      queueName,
			"timeout":         timeout.String(),
		})
	}
}

func rabbitmqClose(ctx context.Context, manager *RabbitMQManager, args []string) (interface{}, error) {
	if len(args) < 1 {
		return nil, util.NewValidationError("rabbitmq close requires a connection name", map[string]interface{}{
			"args_count": len(args),
			"required":   1,
		}).WithAction("rabbitmq").WithStep("close")
	}
	connName := args[0]

	err := manager.CloseConnection(connName)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"status": "disconnected", "connection_name": connName}, nil
}