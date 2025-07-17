package main

import (
	"errors"
	"fmt"

	"github.com/posixenjoyer/learn-pub-sub-starter/internal/gamelogic"
)

func processSpawn(argv []string, state gamelogic.GameState) error {
	err := state.CommandSpawn(argv)
	if err != nil {
		return err
	}
	return nil
}

func processMove(argv []string, state gamelogic.GameState) error {
	move, err := state.CommandMove(argv)
	if err != nil {
		return err
	}
	fmt.Printf("%s moved units to %s\n", move.Player.Username, move.ToLocation)
	return nil
}
func getStatus(argv []string) error {
	return nil
}

func processCmd(argv []string, state *gamelogic.GameState) error {
	var err error
	// check length just in case for sanity

	if len(argv) == 0 {
		return errors.New("no command specified")
	}

	switch argv[0] {
	case "spawn":
		err = processSpawn(argv, *state)
	case "move":
		err = processMove(argv, *state)
	case "status":
		err = getStatus(argv)
	case "spam":
		fmt.Println("Spamming is not allowed yet.")
	}

	return err
}
