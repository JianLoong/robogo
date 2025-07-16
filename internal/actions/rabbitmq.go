package actions

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ action - simplified implementation with proper resource management
func rabbitmqAction(args []interface{}, options map[string]interface{}, vars *common.Variables) (types.ActionResult, error) {
	if len(args) < 3 {
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  "rabbitmq action requires at least 3 arguments: operation, connection_string, queue/exchange",
		}, fmt.Errorf("rabbitmq action requires at least 3 arguments: operation, connection_string, queue/exchange")
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  fmt.Sprintf("failed to connect to RabbitMQ: %v", err),
		}, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	defer conn.Close()

	if conn.IsClosed() {
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  "RabbitMQ connection closed immediately",
		}, fmt.Errorf("RabbitMQ connection closed immediately")
	}

	ch, err := conn.Channel()
	if err != nil {
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  fmt.Sprintf("failed to open channel: %v", err),
		}, fmt.Errorf("failed to open channel: %w", err)
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
		if len(args) < 4 {
			return types.ActionResult{
				Status: types.ActionStatusError,
				Error:  "rabbitmq publish requires: operation, connection_string, queue/exchange, message",
			}, fmt.Errorf("rabbitmq publish requires: operation, connection_string, queue/exchange, message")
		}
		queueOrExchange := fmt.Sprintf("%v", args[2])
		message := fmt.Sprintf("%v", args[3])

		err := ch.PublishWithContext(ctx,
			"",              // exchange
			queueOrExchange, // routing key (queue)
			false,           // mandatory
			false,           // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(message),
			},
		)
		if err != nil {
			return types.ActionResult{
				Status: types.ActionStatusError,
				Error:  fmt.Sprintf("failed to publish message: %v", err),
			}, fmt.Errorf("failed to publish message: %w", err)
		}
		return types.ActionResult{
			Status: types.ActionStatusSuccess,
			Data:   map[string]interface{}{"status": "published"},
		}, nil

	default:
		return types.ActionResult{
			Status: types.ActionStatusError,
			Error:  fmt.Sprintf("unknown rabbitmq operation: %s", operation),
		}, fmt.Errorf("unknown rabbitmq operation: %s", operation)
	}
}
