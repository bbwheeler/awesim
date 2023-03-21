package action_deciders

import "github.com/bbwheeler/awesim/awesim_core/core"

const ActionTypeDoNothing string = "DoNothing"

type DoNothing struct {
	dao core.EntityDao
	timeline *core.Timeline
}

func (dn *DoNothing)DecideActionForActor(actor *core.Actor) (*core.Action, error) {
	return core.NewAction(actor, 1, dn.dao)
}
