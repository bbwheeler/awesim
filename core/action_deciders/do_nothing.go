package action_deciders

import "github.com/bbwheeler/awesim/core"

const ActionTypeDoNothing string = "DoNothing"

type DoNothing struct {
	entityStore core.EntityStore
	timeline    *core.Timeline
}

func NewDoNothingDecider(entityStore core.EntityStore, timeline *core.Timeline) *DoNothing {
	return &DoNothing{
		entityStore: entityStore,
		timeline:    timeline,
	}
}

func (dn *DoNothing) DecideActionForActor(actorID string) (string, error) {
	newAction, err := core.NewAction(actorID, 1, dn.entityStore)
	if err != nil {
		return "", err
	}
	return newAction.GetID(), nil
}
