package core

import "fmt"

type Actor struct {
	Entity
}

func NewActor(dao EntityDao) *Actor {
	entity := NewEntity(dao)
	entity.SetAttribute(IsActorAttribute, true)
	return AsActor(entity)
}

func AsActor(e *Entity) *Actor {
	return &Actor{
		Entity: *e,
	}
}


func (a *Actor) GetNextAction() (*Action, error) {
	actionIDs, err := a.dao.GetEntitiesWithAttribute(actionInvoker, a.GetID())
	if err != nil {
		return nil, err
	}
	if len(actionIDs) > 1 {
		return nil, fmt.Errorf("expected 1 action for invoker but got %v", len(actionIDs))
	}
	if len(actionIDs) < 1 {
		return nil, nil
	}
	return AsAction(GetEntity(actionIDs[0], a.dao)), nil
}
