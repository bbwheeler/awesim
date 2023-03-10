package main

import "github.com/bbwheeler/awesim/awesim_core/engine"
import "github.com/bbwheeler/awesim/awesim_core/dao"
import "github.com/bbwheeler/awesim/awesim_core/action_deciders"
import "github.com/bbwheeler/awesim/awesim_core/action_resolvers"
import "github.com/bbwheeler/awesim/awesim_core/core"

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
	actionDecider := &action_deciders.Random{}
	timeline := core.Timeline{}
	actionResolver := &action_resolvers.Incomplete{}
	engine := engine.New(entityDao, actionDecider, timeline, actionResolver)
	game := NewGame(engine)
	game.Run()
}