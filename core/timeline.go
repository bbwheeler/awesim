package core

import (
	"fmt"
	"math"
)

const timelineEntityID string = "ENTITY_TIMELINE"
const currentTickAttribute string = "CURRENT_TICK"
const actionStartTickAttribute string = "ACTION_START_TICK"

const ErrNoPendingAction = errString("no pending action found")

type Tick = int64

type TimedEntityStore interface {
	EntityStore
	GetDuration(entityID string) (Tick, error)
}

type Timeline struct {
	store TimedEntityStore
}

func NewTimeline(entityStore TimedEntityStore) *Timeline {
	return &Timeline{
		store: entityStore,
	}
}

func (t *Timeline) AddActions(actionIDs []string) error {
	for _, actionID := range actionIDs {
		err := t.AddAction(actionID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Timeline) AddAction(actionID string) error {
	tick, err := t.GetCurrentTick()
	if err != nil {
		return err
	}

	return t.store.SetAttribute(actionID, actionStartTickAttribute, tick)
}

func (t *Timeline) RemoveAction(action *Entity) error {
	return action.RemoveAttribute(actionStartTickAttribute)
}

func (t *Timeline) GetCurrentTick() (Tick, error) {
	timelineEntities, err := t.store.GetEntitiesWithAttributeType(currentTickAttribute)
	if err != nil {
		return 0, err
	}
	if len(timelineEntities) > 1 {
		return 0, fmt.Errorf("expected 1 or 0 timelines but got %v", len(timelineEntities))
	}
	if len(timelineEntities) < 1 {
		return 0, fmt.Errorf("no timeline found, no current tick set")
	}
	currentTick, err := t.store.GetAttribute(timelineEntities[0], currentTickAttribute)
	if err != nil {
		return 0, err
	}
	if currentTick == nil {
		return 0, fmt.Errorf("no current tick")
	}

	return ToTick(currentTick)
}

func (t *Timeline) GetNextActionUpTo(maxTick Tick) (string, error) {
	actionID, err := t.GetPendingActionID()
	if err != nil {
		return "", err
	}

	endTick, err := t.GetEndTickOfAction(actionID)
	if err != nil {
		return "", err
	}

	if endTick > maxTick {
		return "", ErrNoPendingAction
	}

	return actionID, nil
}

func (t *Timeline) GetPendingActionID() (string, error) {
	var earliestEndTick Tick
	var earliestActionID string
	tick, err := t.GetCurrentTick()
	if err != nil {
		return "", err
	}
	actionIDs, err := t.getActions()
	if err != nil {
		return "", err
	}
	for _, actionID := range actionIDs {
		endTick, err := t.GetEndTickOfAction(actionID)
		if err != nil {
			return "", err
		}
		if endTick <= tick && (earliestEndTick <= 0 || endTick < earliestEndTick) {
			earliestEndTick = endTick
			earliestActionID = actionID
		}
	}

	if earliestActionID != "" {
		return earliestActionID, nil
	}

	return "", ErrNoPendingAction
}

func (t *Timeline) GetStartTickOfAction(actionID string) (Tick, error) {
	startTick, err := t.store.GetAttribute(actionID, actionStartTickAttribute)
	if err != nil {
		return 0, err
	}

	return ToTick(startTick)
}
func (t *Timeline) GetEndTickOfAction(actionID string) (Tick, error) {
	startTick, err := t.GetStartTickOfAction(actionID)
	if err != nil {
		return 0, err
	}

	duration, err := t.store.GetDuration(actionID)
	if err != nil {
		return 0, err
	}
	return startTick + duration, nil
}

func (t *Timeline) SetCurrentTick(tick Tick) error {
	return t.store.SetAttribute(timelineEntityID, currentTickAttribute, tick)
}

func indexOfAction(action *Entity, actions []*Entity) (int, error) {
	for index, a := range actions {
		if a.GetID() == action.GetID() {
			return index, nil
		}
	}
	return 0, fmt.Errorf("action %v not found in action array %v", action, actions)
}

func (t *Timeline) getActions() ([]string, error) {
	return t.store.GetEntitiesWithAttributeType(actionStartTickAttribute)
}

func ToTick(v interface{}) (Tick, error) {
	switch n := v.(type) {
	case int64:
		return n, nil
	case int:
		return int64(n), nil
	case int32:
		return int64(n), nil
	case int16:
		return int64(n), nil
	case int8:
		return int64(n), nil
	case uint:
		return int64(n), nil
	case uint64:
		return int64(n), nil
	case uint32:
		return int64(n), nil
	case float64:
		if n != math.Trunc(n) {
			return 0, fmt.Errorf("tick must be a whole number, received %v", n)
		}
		return int64(n), nil
	case float32:
		return int64(n), nil
	default:
		return 0, fmt.Errorf("expected numeric value, received %T", v)
	}
}
