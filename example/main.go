package main

import (
	"fmt"

	"github.com/bbwheeler/awesim/core"
	actionproviders "github.com/bbwheeler/awesim/core/action_providers"
	"github.com/bbwheeler/awesim/core/storage"
	"github.com/bbwheeler/awesim/engine"
)

type Game struct {
	engine *engine.Engine
}

func NewGame(engine *engine.Engine) *Game {
	return &Game{
		engine: engine,
	}
}

func main() {
	entityMapStore := storage.NewEntityMapStore()
	actionStore := core.NewActionStore(entityMapStore)
	actorStore := core.NewActorStore(entityMapStore)
	timeline := core.NewTimeline(actionStore)
	randomActionProvider := &actionproviders.Random{}

	engine := engine.New(entityMapStore, actorStore, actionStore, timeline, randomActionProvider)

	game := NewGame(engine)
	timeline.SetCurrentTick(1)
	fmt.Println("Game Start")
	err := game.engine.ExecuteToTick(10)
	if err != nil {
		fmt.Printf("Game Error: %v\n", err)
	} else {
		fmt.Println("Game End")
	}
}
