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
	timeline := core.NewTimeline(entityMapStore)
	actorProvider := core.NewActorProvider(entityMapStore)

	// actionDecider := action_deciders.NewDoNothingDecider(entityMapStore, timeline)
	// actionResolver := &action_resolvers.NoAction{}
	engine := engine.New(entityMapStore, actorProvider, actionProvider, timeline)

	game := NewGame(engine)
	timeline.SetCurrentTick(1)
	fmt.Println("Game Start")
	err := game.Run()
	if err != nil {
		fmt.Printf("Game Error: %v\n", err)
	} else {
		fmt.Println("Game End")
	}
}
