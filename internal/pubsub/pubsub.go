package pubsub

import (
	"context"
	"encoding/json"
	_ "errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type QueueType int
type AckType int

const (
	Ack = iota
	NackRequeue
	NackDiscard
)

const (
	Durable QueueType = iota
	Transient
)

const consumerStopped = "Stopped consumer, message channel closed."
const gracefulChClose = "Channel Closed properly"

func chanCloseError(err error) string {
	errMsg := fmt.Sprintf("Channel closed as the result of error: %v", err)
	return errMsg
}

func SubscribeJSON[T any](
	con *amqp.Connection,
	exchange, queueName, key string,
	val QueueType,
	handler func(T) AckType) error {

	fmt.Println("Quename = ", queueName)
	aChan, _, err := DeclareBind(con, exchange, queueName, key, val)
	if err != nil {
		return err
	}

	dMsgs, err := aChan.Consume(queueName, "", false, false, false, false, nil)

	if err != nil {
		return err
	}

	statusCh := make(chan string)
	closeCh := make(chan *amqp.Error)
	aChan.NotifyClose(closeCh)

	go func() {
		for {
			defer close(statusCh)
			select {
			case msg, ok := <-dMsgs:
				rawMsg := new(T)
				if !ok {
					statusCh <- consumerStopped
					return
				}

				fmt.Printf("RAW JSON: %s\n", msg.Body)
				err := json.Unmarshal(msg.Body, rawMsg)
				if err != nil {
					statusCh <- fmt.Sprintf(
						"Error unmarshalling JSON: (%s)",
						err)
					return
				}

				ackRes := handler(*rawMsg)

				switch ackRes {
				case Ack:
					msg.Ack(false)
				case NackDiscard:
					msg.Nack(false, false)
				case NackRequeue:
					msg.Nack(false, true)
				}
			case err := <-closeCh:
				if err != nil {
					statusCh <- chanCloseError(err)
				} else {
					statusCh <- gracefulChClose
				}
				return
			}
		}
	}()

	return nil
}

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
	fmt.Printf("RAW JSON: %v\n", string(valJSON))
	err = ch.PublishWithContext(ctx, exchange, key, false, false, publishing)
	if err != nil {
		fmt.Printf("Error publishing: %v\n", err)
		return err
	}

	return nil
}

func DeclareBind(con *amqp.Connection, exchange, queueName, key string, queueType QueueType) (*amqp.Channel, *amqp.Queue, error) {
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

	if err != nil {
		return nil, nil, err
	}

	aTable := make(amqp.Table)

	aTable["x-dead-letter-exchange"] = "peril_dlx"

	queue, err := ch.QueueDeclare(queueName, durable, !durable, !durable, !durable, aTable)
	if err != nil {
		return nil, nil, err
	}

	ch.QueueBind(queueName, key, exchange, !durable, nil)
	return ch, &queue, nil
}
