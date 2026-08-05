package core

import "fmt"

const actionInvoker string = "ACTION_INVOKER"
const ActionDuration string = "ACTION_DURATION"

type ActionStore struct {
	EntityStore
}

func NewActionStore(entityStore EntityStore) *ActionStore {
	return &ActionStore{
		EntityStore: entityStore,
	}
}

func (a *ActionStore) NewAction(invokerID string, duration Tick) (*Entity, error) {
	entity, err := a.NewEntity()
	if err != nil {
		return nil, err
	}

	err = entity.SetAttribute(actionInvoker, invokerID)
	if err != nil {
		return nil, err
	}

	err = entity.SetAttribute(ActionDuration, duration)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (a *ActionStore) GetInvoker(actionID string) (string, error) {
	invokerID, err := a.GetAttribute(actionID, actionInvoker)
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve invoker for action %v: %w", actionID, err)
	}
	if invokerID == nil {
		return "", fmt.Errorf("Action %v has no invoker", actionID)
	}
	return invokerID.(string), nil
}

func (a *ActionStore) GetDuration(actionID string) (Tick, error) {
	duration, err := a.GetAttribute(actionID, ActionDuration)
	if err != nil {
		return 0, err
	}
	durationTick, err := ToTick(duration)

	return durationTick, err
}
