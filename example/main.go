package main

import "fmt"

import "github.com/bbwheeler/awesim/engine"
import "github.com/bbwheeler/awesim/mapstore"
import "github.com/bbwheeler/awesim/core/action_deciders"
import "github.com/bbwheeler/awesim/core/action_resolvers"
import "github.com/bbwheeler/awesim/core"

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
	entityDao := dao.NewEntityDaoMapImpl()
	timeline := core.NewTimeline(entityDao)
	actionDecider := action_deciders.NewDoNothingDecider(entityDao, timeline)
	actionResolver := &action_resolvers.NoAction{}
	engine := engine.New(entityDao, actionDecider, timeline, actionResolver)
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