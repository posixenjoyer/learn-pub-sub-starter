package main

import "fmt"
import (
	gamelogic "github.com/posixenjoyer/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/posixenjoyer/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/posixenjoyer/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"time"
)

type AckType = pubsub.AckType
type ConContext[T any] struct {
	ch   *amqp.Channel
	data T
}

const (
	Ack = iota
	NackRequeue
	NackDiscard
)

func ProcessMakeWar(ctx ConContext[gamelogic.ArmyMove], gs *gamelogic.GameState) AckType {
	move := ctx.data
	warKey := routing.WarRecognitionsPrefix + "." + move.Player.Username
	rw := gamelogic.RecognitionOfWar{
		Attacker: ctx.data.Player,
		Defender: gs.GetPlayerSnap(),
	}

	err := pubsub.PublishJSON(
		ctx.ch,
		routing.ExchangePerilTopic,
		warKey,
		rw)
	if err != nil {
		fmt.Println("Publishing failed: ", err)
		return pubsub.NackRequeue
	}
	return pubsub.Ack
}

func publishWarGob(ctx ConContext[gamelogic.GameState], msg string) AckType {
	exch := routing.ExchangePerilTopic
	key := routing.GameLogSlug + "." + ctx.data.Player.Username
	gameLog := routing.GameLog{
		CurrentTime: time.Now(),
		Message:     msg,
		Username:    ctx.data.Player.Username,
	}

	err := pubsub.PublishGob(ctx.ch, exch, key, gameLog)
	if err != nil {
		return NackRequeue
	}
	return Ack
}

func handleWar(ctx ConContext[gamelogic.GameState]) func(gamelogic.RecognitionOfWar) AckType {
	return func(war gamelogic.RecognitionOfWar) AckType {
		defer fmt.Print("> ")
		outcome, winner, loser := ctx.data.HandleWar(war)
		switch outcome {
		case gamelogic.WarOutcomeNotInvolved:
			fmt.Println("Not Involved!!?")
			return NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			msg := fmt.Sprintf("%s won a war against %s", winner, loser)
			return publishWarGob(ctx, msg)
		case gamelogic.WarOutcomeYouWon:
			msg := fmt.Sprintf("%s won a war against %s", winner, loser)
			return publishWarGob(ctx, msg)
		case gamelogic.WarOutcomeDraw:
			msg := fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
			return publishWarGob(ctx, msg)
		default:
			fmt.Println("We got a weird war error!")
			return NackDiscard
		}
	}
}

func handleMove(ctx ConContext[gamelogic.GameState]) func(gamelogic.ArmyMove) AckType {
	return func(move gamelogic.ArmyMove) AckType {
		defer fmt.Print("> ")
		moveResult := ctx.data.HandleMove(move)

		switch moveResult {
		case gamelogic.MoveOutcomeMakeWar:
			context := ConContext[gamelogic.ArmyMove]{
				ch:   ctx.ch,
				data: move,
			}
			return ProcessMakeWar(context, &ctx.data)
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlePause(state *gamelogic.GameState) func(routing.PlayingState) AckType {
	return func(ps routing.PlayingState) AckType {
		defer fmt.Print("> ")
		state.HandlePause(ps)
		return pubsub.Ack
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

	/*
		Intitial pause message:

		var message routing.PlayingState
		message.IsPaused = true
		err = pubsub.PublishJSON(rabbitChan, string(routing.ExchangePerilDirect), string(routing.PauseKey), message)
		if err != nil {
			fmt.Printf("Error publishing msg: %v\n", err)
			ampqConnection.Close()
			os.Exit(1)
		}
	*/

	queueName := string(routing.PauseKey) + "." + user
	gameState := gamelogic.NewGameState(user)
	ctx := ConContext[gamelogic.GameState]{
		ch:   rabbitChan,
		data: *gameState,
	}

	err = pubsub.SubscribeJSON(ampqConnection,
		routing.ExchangePerilDirect,
		queueName,
		routing.PauseKey,
		pubsub.Transient,
		handlePause(gameState))

	if err != nil {
		fmt.Println("Error Subscribing: ", err)
	}

	queueName = routing.ArmyMovesPrefix + "." + user
	err = pubsub.SubscribeJSON[gamelogic.ArmyMove](ampqConnection,
		routing.ExchangePerilTopic,
		queueName,
		routing.ArmyMovesWC,
		pubsub.Transient,
		handleMove(ctx))

	queueName = routing.WarRecognitionsPrefix
	warKey := routing.WarWC
	err = pubsub.SubscribeJSON[gamelogic.RecognitionOfWar](ampqConnection,
		routing.ExchangePerilTopic,
		queueName,
		warKey,
		pubsub.Durable,
		handleWar(ctx))

	if err != nil {
		fmt.Println("Error subscribing: ", err)
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

		err := processCmd(rabbitChan, input, gameState)
		if err != nil {
			fmt.Println(err)
		}
	}
	ampqConnection.Close()
	os.Exit(0)
}
