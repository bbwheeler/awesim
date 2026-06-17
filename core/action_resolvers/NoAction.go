package action_resolvers

import "github.com/bbwheeler/awesim/core"

type NoAction struct {
}

func (ar *NoAction) ResolveAction(actionID string, entityStore core.EntityStore) (bool, error) {
	action := core.GetAction(actionID, entityStore)
	action.FinishAction()
	return false, nil
}
