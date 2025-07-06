package pubsub

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	valJSON, err := json.Marshal(val)
	if err != nil {
		fmt.Printf("error: val (%v) couldn't be marshalled to JSON!, err: %v\n", val, err)
		return err
	}
	ctx := context.Background()
	var publishing amqp.Publishing
	publishing.ContentType = "application/json"
	publishing.Body = valJSON
	err = ch.PublishWithContext(ctx, exchange, key, false, false, publishing)
	if err != nil {
		fmt.Printf("Error publishing: %v\n", err)
		return err
	}

	return nil
}
