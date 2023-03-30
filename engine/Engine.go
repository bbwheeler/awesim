package engine

import "github.com/bbwheeler/awesim/core"
import "fmt"

type Engine struct {
	entityDao core.EntityDao
	actionDecider core.ActionDecider
	timeline *core.Timeline
	actionResolver core.ActionResolver	
}

func New(dao core.EntityDao, actionDecider core.ActionDecider, timeline *core.Timeline, actionResolver core.ActionResolver) *Engine {
	return &Engine{
		entityDao: dao,
		actionDecider: actionDecider,
		timeline: timeline,
		actionResolver: actionResolver,
	}
}

func (e *Engine) RunOneTurn() error {
	currentTick, err := e.timeline.GetCurrentTick()
	if err != nil {
		return err
	}
	if currentTick <= 0 {
		return fmt.Errorf("current tick must be non-zero and positive")
	} 
	err = e.ExecuteToCurrentTick()
	if err != nil {
		return err
	}
	err = e.timeline.SetCurrentTick(currentTick+1)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Run() error {
	var ended bool
	for !ended {
		err := e.RunOneTurn()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) ExecuteToCurrentTick() error {
	for {
		err := e.DecideActions()
		if err != nil {
			return err
		}
		firstAction, err := e.timeline.GetPendingAction()
		if err != nil {
			return err
		}
		if firstAction == nil {
			return nil
		}
		err = e.ResolveAction(firstAction)
		if err != nil {
			return err
		}
	}
}

func (e *Engine) GetAllActors() ([]*core.Actor, error) {
	entities, err := e.entityDao.GetEntitiesWithAttribute(core.IsActorAttribute, true)
	if err != nil {
		return nil, err
	}
	
	var actors []*core.Actor
	for _, entity := range entities {
		actors = append(actors, core.AsActor(entity))
	}
	return actors, nil
}

func (e *Engine) GetActorsThatNeedActions() ([]*core.Actor, error) {
	allActors, err := e.GetAllActors()
	if err != nil {
		return nil, err
	}
	var actorsNeedingActions []*core.Actor
	for _, actor := range allActors {
		if action, err := actor.GetNextAction(); err == nil && action != nil {
			actorsNeedingActions = append(actorsNeedingActions, actor)
		} else if err != nil {
			return nil, err
		}
	}
	return actorsNeedingActions, nil
}

func (e *Engine) DecideActions() error {
	actorsNeedingActions, err := e.GetActorsThatNeedActions()
	if err != nil {
		return err
	}

	var newActions []*core.Action
	for _, actor := range actorsNeedingActions {
		action, err := e.actionDecider.DecideActionForActor(actor)
		if err != nil {
			return err
		}
		newActions = append(newActions, action)
	}
	return e.timeline.AddActions(newActions)

}

func (e *Engine) ResolveAction(action *core.Action) error {
	resolved, err := e.actionResolver.ResolveAction(action)
	if err != nil {
		return err
	}
	if !resolved {
		return fmt.Errorf("action %v was not resolved", action)
	}
	return e.entityDao.RemoveEntity(action.Entity.GetID())
}