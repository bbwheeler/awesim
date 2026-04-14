package core

import "fmt"

const timelineEntityID string = "ENTITY_TIMELINE"
const currentTickAttribute string = "CURRENT_TICK"
const actionStartTickAttribute string = "ACTION_START_TICK"

const ErrNoPendingAction = errString("no pending action found")

type Timeline struct {
	store EntityStore
}

func NewTimeline(entityStore EntityStore) *Timeline {
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
	action := GetAction(actionID, t.store)
	return action.SetAttribute(actionStartTickAttribute, tick)
}

func (t *Timeline) RemoveAction(action *Action) error {
	return action.RemoveAttribute(actionStartTickAttribute)
}

func (t *Timeline) GetCurrentTick() (int64, error) {
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
	return currentTick.(int64), err
}

func (t *Timeline) GetPendingActionID() (string, error) {
	var earliestEndTick int64
	var earliestAction *Action
	tick, err := t.GetCurrentTick()
	if err != nil {
		return "", err
	}
	actionIDs, err := getActions(t.store)
	if err != nil {
		return "", err
	}
	for _, actionID := range actionIDs {
		action := GetAction(actionID, t.store)
		endTick, err := t.GetEndTickOfAction(action)
		if err != nil {
			return "", err
		}
		if endTick <= tick && (earliestEndTick <= 0 || endTick < earliestEndTick) {
			earliestEndTick = endTick
			earliestAction = action
		}
	}

	if earliestAction != nil {
		return earliestAction.GetID(), nil
	}

	return "", ErrNoPendingAction
}

func (t *Timeline) GetStartTickOfAction(a *Action) (int64, error) {
	startTick, err := a.GetAttribute(actionStartTickAttribute)
	if startTick == nil {
		return 0, fmt.Errorf("Action %v has no start tick", a)
	}
	return startTick.(int64), err
}
func (t *Timeline) GetEndTickOfAction(a *Action) (int64, error) {
	startTick, err := t.GetStartTickOfAction(a)
	if err != nil {
		return 0, err
	}
	duration, err := a.GetDuration()
	if err != nil {
		return 0, err
	}
	return startTick + duration, nil
}

func (t *Timeline) SetCurrentTick(tick int64) error {
	return t.store.SetAttribute(timelineEntityID, currentTickAttribute, tick)
}

func indexOfAction(action *Action, actions []*Action) (int, error) {
	for index, a := range actions {
		if a == action {
			return index, nil
		}
	}
	return 0, fmt.Errorf("action %v not found in action array %v", action, actions)
}

func getActions(store EntityStore) ([]string, error) {
	return store.GetEntitiesWithAttributeType(actionStartTickAttribute)
}
