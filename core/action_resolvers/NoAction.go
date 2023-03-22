package action_resolvers

import "github.com/bbwheeler/awesim/core"

type NoAction struct {
}

func (ar *NoAction)ResolveAction(action *core.Action) (bool, error) {
	return false, nil
}
