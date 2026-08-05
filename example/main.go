package main

import (
	"fmt"

	"github.com/bbwheeler/awesim/core"
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
	actionHoldicator := &Nothing{}

	engine := engine.New(entityMapStore, actorStore, timeline, actionHoldicator, actionHoldicator)

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
