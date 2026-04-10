package core

import "fmt"

const actionInvoker string = "ACTION_INVOKER"
const actionDuration string = "ACTION_DURATION"

type Action struct {
	Entity
}

func NewAction(invoker *Actor, duration int64, store EntityStore) (*Action, error) {
	entity := NewEntity(invoker.store)
	entity.SetAttribute(actionInvoker, invoker.GetID())
	entity.SetAttribute(actionDuration, duration)
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
		return "", fmt.Errorf("Action %v has no invoker")
	}
	return invokerID.(string), nil
}

func (a *Action) GetDuration() (int64, error) {
	duration, err := a.GetAttribute(actionDuration)
	return duration.(int64), err
}

func (a *Action) FinishAction() error {
	return a.store.RemoveAttribute(a.GetID(), actionStartTickAttribute)
}

func asAction(e *Entity) *Action {
	return &Action{
		Entity: *e,
	}
}
