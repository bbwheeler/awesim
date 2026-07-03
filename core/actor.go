package core

import (
	"fmt"
)

const IsActorAttribute string = "IsActor"

const ErrActionNotFoundForActor = errString("action not found for actor")

type Actor struct {
	Entity
}

type ActorProvider struct {
	entityProvider EntityProvider
}
type EntityProvider interface {
	GetEntitiesWithAttribute(entityID string, attribute any) ([]string, error)
	GetEntity(entityID string) *Entity
}

func NewActorProvider(entityProvider EntityProvider) *ActorProvider {
	return &ActorProvider{
		entityProvider: entityProvider,
	}
}

func NewActor(store EntityStore) *Actor {
	entity := NewEntity(store)
	entity.SetAttribute(IsActorAttribute, true)
	return asActor(entity)
}

func (p *ActorProvider) GetAllActorIDs() ([]string, error) {
	return p.entityProvider.GetEntitiesWithAttribute(IsActorAttribute, true)
}

func (p *ActorProvider) GetActor(actorID string) *Actor {
	entity := p.entityProvider.GetEntity(actorID)
	return asActor(entity)
}

func (a *Actor) GetNextAction() (string, error) {
	actionIDs, err := a.store.GetEntitiesWithAttribute(actionInvoker, a.GetID())
	if err != nil {
		return "", err
	}
	if len(actionIDs) > 1 {
		return "", fmt.Errorf("expected 1 action for invoker but got %v", len(actionIDs))
	}
	if len(actionIDs) < 1 {
		return "", ErrActionNotFoundForActor
	}
	return actionIDs[0], nil
}

func (a *Actor) ProvideNextAction() (string, error) {
	panic("unimplemented")
}

func asActor(e *Entity) *Actor {
	return &Actor{
		Entity: *e,
	}
}
