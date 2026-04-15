package main

import (
	"fmt"

	"github.com/bbwheeler/awesim/core"
	"github.com/bbwheeler/awesim/core/action_deciders"
	"github.com/bbwheeler/awesim/core/action_resolvers"
	"github.com/bbwheeler/awesim/engine"
	"github.com/bbwheeler/awesim/mapstore"
)

type Game struct {
	engine *engine.Engine
}

func NewGame(engine *engine.Engine) *Game {
	return &Game{
		engine: engine,
	}
}

func (game *Game) Run() error {
	return game.engine.Run()
}

func main() {
	entityMapStore := mapstore.NewEntityMapStore()
	timeline := core.NewTimeline(entityMapStore)
	actionDecider := action_deciders.NewDoNothingDecider(entityMapStore, timeline)
	actionResolver := &action_resolvers.NoAction{}
	engine := engine.New(entityMapStore, actionDecider, timeline, actionResolver)

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
