package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQ action - simplified implementation with proper resource management
func rabbitmqAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	if len(args) < 3 {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "RABBITMQ_MISSING_ARGS").
			WithTemplate("rabbitmq action requires at least 3 arguments: operation, connection_string, queue/exchange").
			Build()
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "RABBITMQ_CONNECTION_FAILED").
			WithTemplate("failed to connect to RabbitMQ: %v").
			WithContext("connection_string", connectionString).
			WithContext("error", err.Error()).
			Build(err)
	}
	defer conn.Close()

	if conn.IsClosed() {
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "RABBITMQ_CONNECTION_CLOSED").
			WithTemplate("RabbitMQ connection closed immediately").
			WithContext("connection_string", connectionString).
			Build()
	}

	ch, err := conn.Channel()
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryNetwork, "RABBITMQ_CHANNEL_FAILED").
			WithTemplate("failed to open channel: %v").
			WithContext("connection_string", connectionString).
			WithContext("error", err.Error()).
			Build(err)
	}
	defer func() {
		if closeErr := ch.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close RabbitMQ channel: %v\n", closeErr)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultMessagingTimeout)
	defer cancel()

	switch operation {
	case constants.OperationPublish:
		if len(args) < 4 {
			return types.NewErrorBuilder(types.ErrorCategoryValidation, "RABBITMQ_PUBLISH_MISSING_ARGS").
				WithTemplate("rabbitmq publish requires: operation, connection_string, queue/exchange, message").
				Build()
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
			return types.NewErrorBuilder(types.ErrorCategoryNetwork, "RABBITMQ_PUBLISH_FAILED").
				WithTemplate("failed to publish message: %v").
				WithContext("connection_string", connectionString).
				WithContext("queue", queueOrExchange).
				WithContext("error", err.Error()).
				Build(err)
		}
		return types.ActionResult{
			Status: types.ActionStatusPassed,
			Data:   map[string]any{"status": "published"},
		}

	default:
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "RABBITMQ_UNKNOWN_OPERATION").
			WithTemplate("unknown rabbitmq operation: %s").
			WithContext("operation", operation).
			Build(operation)
	}
}
