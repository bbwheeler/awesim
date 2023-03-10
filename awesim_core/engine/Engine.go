package engine

import "github.com/bbwheeler/awesim/awesim_core/dao"
import "github.com/bbwheeler/awesim/awesim_core/core"

type Engine struct {
	entityDao dao.EntityDao
	actionDecider core.ActionDecider
	timeline core.Timeline
	actionResolver core.ActionResolver
	
}

func New(dao dao.EntityDao, actionDecider core.ActionDecider, timeline core.Timeline, actionResolver core.ActionResolver) *Engine {
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

	var newActions []core.Action
	for _, actor := range actorsNeedingActions {
		action, err := e.actionDecider.GetActionForActor(actor)
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

	err = e.actionResolver.ResolveAction(actionToExecute)
	if err != nil {
		return false, err
	}
	err = e.timeline.RemoveAction(actionToExecute)
	if err != nil {
		return false, err
	}

	return true, nil
}

type actor struct {
	entity core.Entity
}

func (a actor) GetID() string {
	return a.entity.GetID()
}
func (a actor) GetStringAttribute(attribute string) (string, error) {
	return a.entity.GetStringAttribute(attribute)
}
func (a actor) GetNumberAttribute(attribute string) (float64, error) {
	return a.entity.GetNumberAttribute(attribute)
} 
func (a actor) GetIntegerAttribute(attribute string) (int64, error) {
	return a.entity.GetIntegerAttribute(attribute)
}
func (a actor) GetBooleanAttribute(attribute string) (bool, error) {
	return a.entity.GetBooleanAttribute(attribute)
}

func (a actor) GetNextAction() (core.Action, error) {
	// TODO
	return nil, nil
}



func (e *Engine) GetAllActors() ([]core.Actor, error) {
	entities, err := e.entityDao.GetEntitiesWithBooleanAttribute(core.IsActorAttribute, true)
	if err != nil {
		return nil, err
	}

	var actors []core.Actor
	for _, entity := range entities {
		actor := actor{
			entity: entity,
		}
		actors = append(actors, actor)
	}
	return actors, nil
}

func (e *Engine) GetActorsThatNeedActions() ([]core.Actor, error) {
	allActors, err := e.GetAllActors()
	if err != nil {
		return nil, err
	}
	var actorsNeedingActions []core.Actor
	for _, actor := range allActors {
		if action, err := actor.GetNextAction(); err == nil && action != nil {
			actorsNeedingActions = append(actorsNeedingActions, actor)
		} else if err != nil {
			return nil, err
		}
	}
	return actorsNeedingActions, nil
}
