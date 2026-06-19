package action_resolvers

import "github.com/bbwheeler/awesim/core"

type NoAction struct {
	entityStore core.EntityStore
}

func (ar *NoAction) ResolveAction(actionID string) (bool, error) {
	action := core.GetAction(actionID, ar.entityStore)
	action.FinishAction()
	return false, nil
}
