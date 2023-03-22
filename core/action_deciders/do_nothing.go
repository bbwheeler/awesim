package action_deciders

import "github.com/bbwheeler/awesim/core"

const ActionTypeDoNothing string = "DoNothing"

type DoNothing struct {
	dao core.EntityDao
	timeline *core.Timeline
}

func NewDoNothingDecider(dao core.EntityDao, timeline *core.Timeline) *DoNothing {
	return &DoNothing{
		dao: dao,
		timeline: timeline,
	}
}

func (dn *DoNothing)DecideActionForActor(actor *core.Actor) (*core.Action, error) {
	return core.NewAction(actor, 1, dn.dao)
}
