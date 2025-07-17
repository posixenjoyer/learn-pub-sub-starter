package main

import (
	"fmt"
	"os"

	"github.com/posixenjoyer/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/posixenjoyer/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/posixenjoyer/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type commandType int

const (
	Pause = iota
	Resume
)

func publishMsg(ch *amqp.Channel, message routing.PlayingState) {
	err := pubsub.PublishJSON(ch, string(routing.ExchangePerilDirect), string(routing.PauseKey), message)
	if err != nil {
		fmt.Printf("Error publishing msg: %v\n", err)
		os.Exit(1)
	}
}

func processCommand(ch *amqp.Channel, cmdType commandType) {
	var message routing.PlayingState

	if cmdType == Pause {
		message.IsPaused = true
		publishMsg(ch, message)
	}

	if cmdType == Resume {
		message.IsPaused = false
		publishMsg(ch, message)
	}
}

func main() {
	fmt.Println("Starting Peril server...")
	connect := "amqp://guest:guest@localhost:5672/"
	amqpConnection, err := amqp.Dial(connect)

	if err != nil {
		fmt.Printf("Error connecting to AMQP server: %v\n", err)
	}
	defer amqpConnection.Close()

	rabbitChan, err := amqpConnection.Channel()
	if err != nil {
		fmt.Printf("Error setting up message channel: %v\n", err)
	}

	_, _, err = pubsub.DeclareBind(amqpConnection,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PauseKey,
		pubsub.Durable)
	var message routing.PlayingState
	message.IsPaused = true
	err = pubsub.PublishJSON(rabbitChan, routing.ExchangePerilDirect, routing.PauseKey, message)
	if err != nil {
		fmt.Printf("Error publishing msg: %v\n", err)
		os.Exit(1)
	}

	_, _, err = pubsub.DeclareBind(
		amqpConnection,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogKey,
		pubsub.Durable)

	for {
		gamelogic.PrintServerHelp()
		input := gamelogic.GetInput()
		if len(input) < 1 {
			fmt.Println("Please enter a command!!")
			continue
		}

		if input[0] == "quit" {
			fmt.Println("'quit' received... exiting")
			break
		}

		switch input[0] {
		case "pause":
			fmt.Println("Sending 'pause' message...")
			processCommand(rabbitChan, Pause)
		case "resume":
			fmt.Println("Sending 'resume' message...")
			processCommand(rabbitChan, Resume)
		default:
			fmt.Println("Unrecognized command, try again...")
		}
	}

	os.Exit(0)
}
