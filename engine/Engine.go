package engine

import (
	"errors"
	"fmt"

	"github.com/bbwheeler/awesim/core"
)

type Engine struct {
	entityProvider entityProvider
	timeline       timeline
	actorProvider  actorProvider
	actionProvider actionProvider
}

type Actor interface {
	GetNextActionID() (string, error)
	ProvideNextAction() (string, error)
	GetID() string
}

type Action interface {
	Resolve() error
}

type timeline interface {
	GetCurrentTick() (core.Tick, error)
	SetCurrentTick(tick core.Tick) error
	GetPendingActionID() (string, error)
	GetNextActionUpTo(maxTick core.Tick) (string, error)
	AddActions(actionIDs []string) error
}

type actorProvider interface {
	GetActor(actorID string) (Actor, error)
	GetAllActorIDs() ([]string, error)
}

type actionProvider interface {
	GetAction(actionID string) (Action, error)
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

func New(entityProvider entityProvider, actorProvider actorProvider, actionProvider actionProvider, timeline timeline) *Engine {
	return &Engine{
		entityProvider: entityProvider,
		timeline:       timeline,
		actorProvider:  actorProvider,
		actionProvider: actionProvider,
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
		err = e.updateActions()
		if err != nil {
			return err
		}

		firstActionID, err := e.timeline.GetNextActionUpTo(targetTick)
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
	allActorIDs, err := e.actorProvider.GetAllActorIDs()
	if err != nil {
		return nil, err
	}
	var actorsNeedingActions []string
	for _, actorID := range allActorIDs {
		actor, err := e.actorProvider.GetActor(actorID)
		if err != nil {
			return nil, err
		}

		if actionID, err := actor.GetNextActionID(); err == nil && actionID == "" {
			actorsNeedingActions = append(actorsNeedingActions, actorID)
		} else if err != nil {
			return nil, err
		}
	}
	return actorsNeedingActions, nil
}

func (e *Engine) updateActions() error {
	actorsNeedingActions, err := e.getActorsThatNeedActions()
	if err != nil {
		return err
	}

	var newActions []string
	for _, actorID := range actorsNeedingActions {
		actor, err := e.actorProvider.GetActor(actorID)
		if err != nil {
			return err
		}
		actionID, err := actor.ProvideNextAction()
		if err != nil {
			return err
		}
		if actionID != "" {
			newActions = append(newActions, actionID)
		}
	}
	return e.timeline.AddActions(newActions)
}

func (e *Engine) resolveAction(actionID string) error {
	action, err := e.actionProvider.GetAction(actionID)
	if err != nil {
		return err
	}

	err = action.Resolve()
	if err != nil {
		return err
	}

	return e.entityProvider.RemoveEntity(actionID)
}
