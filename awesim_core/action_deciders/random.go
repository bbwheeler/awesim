package action_deciders

import "github.com/bbwheeler/awesim/awesim_core/core"
import (
    "fmt"
    "math/rand"
)

type Random struct {
}

func (ad *Incomplete)GetActionForActor(actor core.Actor) (core.Action, error) {
	types := actor.GetPossibleActionTypes()
	randomIndex := rand.Intn(len(types))
	t := types[randomIndex]
	return t.CreateActionForActor(actor)
}
