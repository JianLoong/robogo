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
		return types.MissingArgsError("rabbitmq", 3, len(args))
	}

	operation := strings.ToLower(fmt.Sprintf("%v", args[0]))
	connectionString := fmt.Sprintf("%v", args[1])

	conn, err := amqp.Dial(connectionString)
	if err != nil {
		return types.ConnectionError("RabbitMQ", err.Error())
	}
	defer conn.Close()

	if conn.IsClosed() {
		return types.ConnectionError("RabbitMQ", "connection closed immediately")
	}

	ch, err := conn.Channel()
	if err != nil {
		return types.RequestError("RabbitMQ channel open", err.Error())
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
			return types.MissingArgsError("rabbitmq publish", 4, len(args))
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
			return types.RequestError(fmt.Sprintf("rabbitmq publish to %s", queueOrExchange), err.Error())
		}
		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data:   map[string]any{"status": "published"},
		}

	default:
		return types.UnknownOperationError("rabbitmq", operation)
	}
}
