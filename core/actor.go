package core

import (
	"fmt"
)

const IsActorAttribute string = "IsActor"

const ErrActionNotFoundForActor = errString("action not found for actor")

type ActorStore struct {
	EntityStore
}

func NewActorStore(entityStore EntityStore) *ActorStore {
	return &ActorStore{
		EntityStore: entityStore,
	}
}

func (s *ActorStore) NewActor() (*Entity, error) {
	entity, err := s.NewEntity()
	if err != nil {
		return nil, err
	}
	err = entity.SetAttribute(IsActorAttribute, true)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (p *ActorStore) GetAllActorIDs() ([]string, error) {
	return p.GetEntitiesWithAttribute(IsActorAttribute, true)
}

func (a *ActorStore) GetNextActionID(actorID string) (string, error) {
	actionIDs, err := a.GetEntitiesWithAttribute(actionInvoker, actorID)
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
