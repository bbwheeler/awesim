package actionproviders

import "github.com/bbwheeler/awesim/core"

const ActionDoNothing string = "ActionDoNothingDoNothing"

type DoNothing struct {
	entityStore core.EntityStore
	timeline    *core.Timeline
}

func DoNothingAllActions() []string {
	return []string{ActionDoNothing}
}

func (dn *DoNothing) AvailableActions(actorID string) ([]string, error) {
	return DoNothingAllActions(), nil
}

func (dn *DoNothing) CreateAction(actionType string, actorID string, additionalParameters map[string]any) (actionID string, err error) {
	duration, ok := additionalParameters[core.ActionDuration]
	if !ok {
		duration = 1
	}
	durationTick, err := core.ToTick(duration)
	if err != nil {
		return "", err
	}

	action, err := core.NewAction(actorID, durationTick, dn.entityStore)
	if err != nil {
		return "", err
	}

	return action.GetID(), nil
}
