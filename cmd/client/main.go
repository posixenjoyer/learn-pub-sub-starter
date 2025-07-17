package main

import "fmt"
import (
	gamelogic "github.com/posixenjoyer/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/posixenjoyer/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/posixenjoyer/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
)

func handlerPause(state *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
		fmt.Printf("Handler called with: %+v\n", ps)
		defer fmt.Print("> ")
		state.HandlePause(ps)
	}
}

func main() {
	fmt.Println("Starting Peril client...")

	connect := "amqp://guest:guest@localhost:5672/"
	ampqConnection, err := amqp.Dial(connect)

	if err != nil {
		fmt.Printf("Error connecting to AMQP server: %v\n", err)
	}
	defer ampqConnection.Close()

	rabbitChan, err := ampqConnection.Channel()
	if err != nil {
		fmt.Printf("Error setting up message channel: %v\n", err)
	}

	user, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Error getting username: %v\n", err)
	}

	var message routing.PlayingState
	message.IsPaused = true
	err = pubsub.PublishJSON(rabbitChan, string(routing.ExchangePerilDirect), string(routing.PauseKey), message)
	if err != nil {
		fmt.Printf("Error publishing msg: %v\n", err)
		os.Exit(1)
	}

	queueName := string(routing.PauseKey) + "." + user
	gameState := gamelogic.NewGameState(user)
	err = pubsub.SubscribeJSON(ampqConnection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState))

	if err != nil {
		fmt.Println("Error Subscribing: ", err)
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			fmt.Println("Oh, a funny guy....")
			continue
		}

		if input[0] == "quit" {
			gamelogic.PrintQuit()
			break
		}

		if gameState.Paused {
			fmt.Println("Sorry, the game is paused!")
			continue
		}

		err := processCmd(input, gameState)
		if err != nil {
			fmt.Println(err)
		}
	}
	os.Exit(0)
}
