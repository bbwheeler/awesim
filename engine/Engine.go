package engine

import (
	"errors"
	"fmt"

	"github.com/bbwheeler/awesim/core"
)

type Engine struct {
	entityStore   entityStore
	actorStore    actorStore
	actionStore   actionStore
	timeline      timeline
	actionCreator actionProvider
}

type timeline interface {
	GetCurrentTick() (core.Tick, error)
	SetCurrentTick(tick core.Tick) error
	GetPendingActionID() (string, error)
	GetNextActionUpTo(maxTick core.Tick) (string, error)
	AddActions(actionIDs []string) error
}
type entityStore interface {
	RemoveEntity(entityID string) error

	GetAttribute(entityId string, attributeId string) (any, error)
	HasAttribute(entityId string, attributeId string) (bool, error)
	SetAttribute(entityId string, attributeId string, value any) error
	RemoveAttribute(entityId string, attributeId string) error

	GetEntitiesWithAttributes(attributes map[string]any) ([]string, error)
	GetEntitiesWithAttributeType(attribute string) ([]string, error)
	GetEntitiesWithAttribute(attribute string, value any) ([]string, error)
}

type actorStore interface {
	GetAllActorIDs() ([]string, error)
	GetNextActionID(actorID string) (string, error)
}
type actionStore interface {
}

type actionProvider interface {
	ProvideNextActionFor(actorID string) (string, error)
}

func New(entityStore entityStore, actorStore actorStore, actionStore actionStore, timeline timeline, actionCreator actionProvider) *Engine {
	return &Engine{
		entityStore:   entityStore,
		actorStore:    actorStore,
		actionStore:   actionStore,
		timeline:      timeline,
		actionCreator: actionCreator,
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
	allActorIDs, err := e.actorStore.GetAllActorIDs()
	if err != nil {
		return nil, err
	}
	var actorsNeedingActions []string
	for _, actorID := range allActorIDs {

		if actionID, err := e.actorStore.GetNextActionID(actorID); err == nil && actionID == "" {
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
		actionID, err := e.actionCreator.ProvideNextActionFor(actorID)
		if err != nil {
			return err
		}
		// actionID, err := e.actorStore.GetNextActionID(actorID)
		// if err != nil {
		// 	return err
		// }

		if actionID != "" {
			newActions = append(newActions, actionID)
		}
	}
	return e.timeline.AddActions(newActions)
}

func (e *Engine) resolveAction(actionID string) error {
	// action := e.actionProvider.GetAction(actionID)

	// err := action.Resolve()
	// if err != nil {
	// 	return err
	// }

	// return e.entityProvider.RemoveEntity(actionID)

	// TODO: This
	return nil
}
