package engine

import (
	"errors"
	"fmt"

	"github.com/bbwheeler/awesim/core"
)

type Engine struct {
	entityProvider entityProvider
	actionDecider  actionDecider
	timeline       timeline
	actionResolver actionResolver
}

type Actor interface {
	GetNextActionID() (string, error)
	GetID() string
}

type timeline interface {
	GetCurrentTick() (core.Tick, error)
	SetCurrentTick(tick core.Tick) error
	GetPendingActionID() (string, error)
	GetPendingActionIDWithMaxTick(maxTick core.Tick) (string, error)
	AddActions(actionIDs []string) error
}

type actorProvider interface {
	GetActor(actorID string) (Actor, error)
}

type entityProvider interface {
	RemoveEntity(entityID string) error

	GetAttribute(entityId string, attributeId string) (any, error)
	HasAttribute(entityId string, attributeId string) (bool, error)
	SetAttribute(entityId string, attributeId string, value any) error
	RemoveAttribute(entityId string, attributeId string) error

	GetEntitiesWithAttributes(attributes map[string]any) ([]string, error)
	GetEntitiesWithAttributeType(attribute string) ([]string, error)
	GetEntitiesWithAttribute(attribute string, value any) ([]string, error)
}

type actionDecider interface {
	DecideActionForActor(actorID string) (string, error)
}

type actionResolver interface {
	ResolveAction(actionID string) (bool, error)
}

func New(entityProvider entityProvider, actionDecider actionDecider, timeline timeline, actionResolver actionResolver) *Engine {
	return &Engine{
		entityProvider: entityProvider,
		actionDecider:  actionDecider,
		timeline:       timeline,
		actionResolver: actionResolver,
	}
}

func (e *Engine) RunCurrentTurn() error {
	currentTick, err := e.timeline.GetCurrentTick()
	if err != nil {
		return err
	}
	if currentTick < 0 {
		return fmt.Errorf("current tick must be positive")
	}

	err = e.ExecuteToTick(currentTick)
	if err != nil {
		return err
	}

	err = e.timeline.SetCurrentTick(currentTick + 1)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Run() error {
	var ended bool
	for !ended {
		err := e.RunCurrentTurn()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) ExecuteToTick(targetTick core.Tick) error {
	currentTick, err := e.timeline.GetCurrentTick()
	if err != nil {
		return err
	}
	if currentTick < 0 {
		return fmt.Errorf("current tick must be positive")
	}
	if currentTick > targetTick {
		return nil
	}

	for {
		err = e.decideActions()
		if err != nil {
			return err
		}
		firstActionID, err := e.timeline.GetPendingActionIDWithMaxTick(targetTick)
		if err != nil {
			if errors.Is(err, core.ErrNoPendingAction) {
				return nil
			}
			return err
		}

		err = e.resolveAction(firstActionID)
		if err != nil {
			return err
		}
	}
}

func (e *Engine) getActorsThatNeedActions() ([]string, error) {
	allActorIDs, err := core.GetAllActors(e.entityProvider)
	if err != nil {
		return nil, err
	}
	var actorsNeedingActions []string
	for _, actorID := range allActorIDs {
		actor := core.GetActor(actorID, e.entityProvider)
		if actionID, err := actor.GetNextAction(); err == nil && actionID == "" {
			actorsNeedingActions = append(actorsNeedingActions, actorID)
		} else if err != nil {
			return nil, err
		}
	}
	return actorsNeedingActions, nil
}

func (e *Engine) decideActions() error {
	actorsNeedingActions, err := e.getActorsThatNeedActions()
	if err != nil {
		return err
	}

	var newActions []string
	for _, actorID := range actorsNeedingActions {
		actionID, err := e.actionDecider.DecideActionForActor(actorID)
		if err != nil {
			return err
		}
		if actionID == "" {
			continue
		}
		newActions = append(newActions, actionID)
	}
	return e.timeline.AddActions(newActions)
}

func (e *Engine) resolveAction(actionID string) error {
	resolved, err := e.actionResolver.ResolveAction(actionID)
	if err != nil {
		return err
	}
	if !resolved {
		return fmt.Errorf("action %v was not resolved", actionID)
	}
	return e.entityProvider.RemoveEntity(actionID)
}
