package pubsub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type queueType int

const (
	Durable queueType = iota
	Transient
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

func DeclareBind(con *amqp.Connection, exchange, queueName, key string, queueType queueType) (*amqp.Channel, *amqp.Queue, error) {
	var durable bool

	switch queueType {
	case Durable:
		durable = true
	case Transient:
		durable = false
	default:
		durable = false
	}

	ch, err := con.Channel()
	defer ch.Close()

	if err != nil {
		return nil, nil, err
	}

	queue, err := ch.QueueDeclare(queueName, durable, !durable, !durable, !durable, nil)
	if err != nil {
		return nil, nil, err
	}

	ch.QueueBind(queueName, key, exchange, !durable, nil)
	return ch, &queue, nil
}
