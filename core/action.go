package core

import "fmt"

const actionInvoker string = "ACTION_INVOKER"
const ActionDuration string = "ACTION_DURATION"

type Action interface {
	Resolve() error
	SetAttribute(attribute string, value any) error
	GetAttribute(attribute string) (any, error)
	GetDuration() (Tick, error)
}

type ActionEntity struct {
	Entity
}

type ActionProvider struct {
	entityProvider EntityProvider
}

func NewActionProvider(entityProvider EntityProvider) *ActionProvider {
	return &ActionProvider{
		entityProvider: entityProvider,
	}
}

func NewAction(invokerID string, duration Tick, store EntityStore) (*ActionEntity, error) {
	entity := NewEntity(store)
	entity.SetAttribute(actionInvoker, invokerID)
	entity.SetAttribute(ActionDuration, duration)
	return &ActionEntity{
		Entity: *entity,
	}, nil
}

func (p *ActionProvider) GetAction(actionID string) Action {
	entity := p.entityProvider.GetEntity(actionID)
	return asAction(entity)
}

func (a *ActionEntity) GetInvoker() (string, error) {
	invokerID, err := a.GetAttribute(actionInvoker)
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve invoker for action %v: %w", a.GetID(), err)
	}
	if invokerID == nil {
		return "", fmt.Errorf("Action %v has no invoker", a.GetID())
	}
	return invokerID.(string), nil
}

func (a *ActionEntity) GetDuration() (Tick, error) {
	duration, err := a.GetAttribute(ActionDuration)
	if err != nil {
		return 0, err
	}
	durationTick, err := ToTick(duration)

	return durationTick, err
}

func (a *ActionEntity) FinishAction() error {
	return a.store.RemoveAttribute(a.GetID(), actionStartTickAttribute)
}

func (a *ActionEntity) Resolve() error {
	// TODO: Implement
	return nil
}

func asAction(e *Entity) *ActionEntity {
	return &ActionEntity{
		Entity: *e,
	}
}
