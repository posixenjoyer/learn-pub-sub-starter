package main

import (
	"fmt"
	pubsub "github.com/posixenjoyer/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/posixenjoyer/learn-pub-sub-starter/internal/routing"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")
	connect := "amqp://guest:guest@localhost:5672/"
	ampqConnection, err := amqp.Dial(connect)

	if err != nil {
		fmt.Printf("Error connecting to AMQP server: %v\n", err)
	}
	defer ampqConnection.Close()

	sigChan := make(chan os.Signal, 1)
	rabbitChan, err := ampqConnection.Channel()
	if err != nil {
		fmt.Printf("Error setting up message channel: %v\n", err)
	}

	var message routing.PlayingState
	message.IsPaused = true
	err = pubsub.PublishJSON(rabbitChan, string(routing.ExchangePerilDirect), string(routing.PauseKey), message)
	if err != nil {
		fmt.Printf("Error publishing msg: %v\n", err)
		os.Exit(1)
	}

	signal.Notify(sigChan, os.Interrupt)
	<-sigChan
	fmt.Println("Received interrupt, exiting...")
	os.Exit(0)
}
