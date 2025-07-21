package main

import (
	"errors"
	"fmt"

	"github.com/posixenjoyer/learn-pub-sub-starter/internal/gamelogic"
	"github.com/posixenjoyer/learn-pub-sub-starter/internal/pubsub"
	"github.com/posixenjoyer/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"strconv"
	"time"
)

func processSpawn(ch *amqp.Channel, argv []string, state gamelogic.GameState) error {
	err := state.CommandSpawn(argv)
	if err != nil {
		return err
	}
	return nil
}

func processSpam(ch *amqp.Channel, argv []string, state gamelogic.GameState) error {
	fmt.Printf("argv[1]: %v\n", argv[1])
	n, err := strconv.Atoi(argv[1])
	if err != nil {
		return err
	}

	for i := 0; i < n; i++ {
		msg := gamelogic.GetMaliciousLog()
		gameLog := routing.GameLog{
			CurrentTime: time.Now(),
			Message:     msg,
			Username:    state.Player.Username,
		}
		err = pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+state.Player.Username, gameLog)
		if err != nil {
			fmt.Println("Publish error: ", err)
		}
	}
	return nil
}

func processMove(ch *amqp.Channel, argv []string, state gamelogic.GameState) error {
	move, err := state.CommandMove(argv)
	if err != nil {
		return err
	}
	fmt.Printf("%s moved to %s\n", move.Player.Username, move.ToLocation)
	key := routing.ArmyMovesPrefix + "." + move.Player.Username
	err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, key, move)
	if err != nil {
		fmt.Println("Move failed: Error: ", err)
	} else {
		fmt.Println("Move published successfully")
	}

	return nil
}
func getStatus(argv []string) error {
	return nil
}

func processCmd(ch *amqp.Channel, args []string, state *gamelogic.GameState) error {
	var err error
	// check length just in case for sanity

	if len(args) == 0 {
		return errors.New("no command specified")
	}

	switch args[0] {
	case "spawn":
		err = processSpawn(ch, args, *state)
	case "move":
		err = processMove(ch, args, *state)
	case "status":
		err = getStatus(args)
	case "spam":
		err = processSpam(ch, args, *state)
	}

	return err
}
