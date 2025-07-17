package main

import "fmt"
import (
	gamelogic "github.com/posixenjoyer/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/posixenjoyer/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/posixenjoyer/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
)

func handleMove(state *gamelogic.GameState) func(gamelogic.ArmyMove) {
	return func(move gamelogic.ArmyMove) {
		defer fmt.Print("> ")
		state.HandleMove(move)
	}
}

func handlePause(state *gamelogic.GameState) func(routing.PlayingState) {
	return func(ps routing.PlayingState) {
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
		handlePause(gameState))

	if err != nil {
		fmt.Println("Error Subscribing: ", err)
	}

	queueName := routing.ArmyMovesPrefix + "." + user
	err = pubsub.SubscribeJSON[gamelogic.ArmyMove](ampqConnection,
		routing.ExchangePerilTopic,
		queueName,
		routing.ArmyMovesWC,
		pubsub.Transient,
		handleMove(gameState))

	if err != nil {
		fmt.Println("Error subscribing: ", err)
	}
	fmt.Println("queueName: ", queueName)
	fmt.Println("routingKey: ", routing.ArmyMovesWC)
	for {
		fmt.Println("How the fuck did I get here?")

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

		err := processCmd(rabbitChan, input, gameState)
		if err != nil {
			fmt.Println(err)
		}
	}
	os.Exit(0)
}
