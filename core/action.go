package core

import "fmt"

const actionInvoker string = "ACTION_INVOKER"
const ActionDuration string = "ACTION_DURATION"

type Action struct {
	Entity
}

func NewAction(invokerID string, duration Tick, store EntityStore) (*Action, error) {
	entity := NewEntity(store)
	entity.SetAttribute(actionInvoker, invokerID)
	entity.SetAttribute(ActionDuration, duration)
	return &Action{
		Entity: *entity,
	}, nil
}

func GetAction(actionID string, entityStore EntityStore) *Action {
	entity := GetEntity(actionID, entityStore)
	return asAction(entity)
}

func (a *Action) GetInvoker() (string, error) {
	invokerID, err := a.GetAttribute(actionInvoker)
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve invoker for action %v: %w", a.GetID(), err)
	}
	if invokerID == nil {
		return "", fmt.Errorf("Action %v has no invoker", a.GetID())
	}
	return invokerID.(string), nil
}

func (a *Action) GetDuration() (Tick, error) {
	duration, err := a.GetAttribute(ActionDuration)
	if err != nil {
		return 0, err
	}
	durationTick, err := ToTick(duration)

	return durationTick, err
}

func (a *Action) FinishAction() error {
	return a.store.RemoveAttribute(a.GetID(), actionStartTickAttribute)
}

func asAction(e *Entity) *Action {
	return &Action{
		Entity: *e,
	}
}
