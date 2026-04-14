package core

import "fmt"

const IsActorAttribute string = "IsActor"

const ErrActionNotFoundForActor = errString("action not found for actor")

type Actor struct {
	Entity
}

type EntityProvider interface {
	GetEntitiesWithAttribute(entityID string, attribute any) ([]string, error)
}

func NewActor(store EntityStore) *Actor {
	entity := NewEntity(store)
	entity.SetAttribute(IsActorAttribute, true)
	return asActor(entity)
}

func GetAllActors(entityProvider EntityProvider) ([]string, error) {
	return entityProvider.GetEntitiesWithAttribute(IsActorAttribute, true)
}

func GetActor(actorID string, entityStore EntityStore) *Actor {
	entity := GetEntity(actorID, entityStore)
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

func asActor(e *Entity) *Actor {
	return &Actor{
		Entity: *e,
	}
}
