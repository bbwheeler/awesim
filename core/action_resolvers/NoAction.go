package action_resolvers

type NoAction struct {
}

func (ar *NoAction) ResolveAction(actionID string) (bool, error) {
	return false, nil
}
