package engine

import (
	"errors"
	"fmt"

	"github.com/bbwheeler/awesim/core"
)

type Engine struct {
	entityStore    entityStore
	actorStore     actorStore
	timeline       timeline
	actionProvider actionProvider
	actionResolver actionResolver
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
}

type actorStore interface {
	GetAllActorIDs() ([]string, error)
	GetNextActionID(actorID string) (string, error)
}

type actionProvider interface {
	ProvideNextActionFor(actorID string) (string, error)
}

type actionResolver interface {
	ResolveAction(actionID string) error
}

func New(entityStore entityStore, actorStore actorStore, timeline timeline, actionProvider actionProvider, actionResolver actionResolver) *Engine {
	return &Engine{
		entityStore:    entityStore,
		actorStore:     actorStore,
		timeline:       timeline,
		actionProvider: actionProvider,
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
		actionID, err := e.actionProvider.ProvideNextActionFor(actorID)
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
	err := e.actionResolver.ResolveAction(actionID)
	if err != nil {
		return err
	}

	return e.entityStore.RemoveEntity(actionID)
}
