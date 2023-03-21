package engine

import "github.com/bbwheeler/awesim/awesim_core/core"
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

func (e *Engine) Run() error {
	var ended bool
	var err error
	for !ended {
		ended, err = e.ExecuteOneTurn()
		if err != nil {
			return err
		} 
	}
	return nil
}

func (e *Engine) ExecuteOneTurn() (bool, error) {
	actorsNeedingActions, err := e.GetActorsThatNeedActions()
	if err != nil {
		return false, err
	}

	var newActions []*core.Action
	for _, actor := range actorsNeedingActions {
		action, err := e.actionDecider.DecideActionForActor(actor)
		if err != nil {
			return false, err
		}
		newActions = append(newActions, action)
	}
	e.timeline.AddActions(newActions)
	actionToExecute, err := e.timeline.GetFirstPendingAction()
	if err != nil {
		return false, err
	}
	if actionToExecute == nil {
		return false, nil
	}
	actionStart, err := e.timeline.GetStartTickOfAction(actionToExecute)
	if err != nil {
		return false, err
	}
	currentTick, err := e.timeline.GetCurrentTick()
	if actionStart <= currentTick {

		resolved, err := e.actionResolver.ResolveAction(actionToExecute)
		if err != nil {
			return false, err
		}
		if !resolved {
			return false, fmt.Errorf("action %v was not resolved", actionToExecute)
		}
		e.entityDao.RemoveEntity(&actionToExecute.Entity)
	} else {
		err = e.timeline.SetCurrentTick(currentTick+1)
		if err != nil {
			return false, err
		}
	}
	return true, nil
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
