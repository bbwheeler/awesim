package core

import "fmt"

type Actor struct {
	dao EntityDao
	Entity
}

func AsActor(e *Entity) *Actor {
	return &Actor{
		dao: e.dao,
		Entity: *e,
	}
}


func (a *Actor) GetNextAction() (*Action, error) {
	actions, err := a.dao.GetEntitiesWithAttribute(actionInvoker, a.GetID())
	if err != nil {
		return nil, err
	}
	if len(actions) > 1 {
		return nil, fmt.Errorf("expected 1 action for invoker but got %v", len(actions))
	}
	return AsAction(actions[0]), nil
}
