package core

import "fmt"

const timelineEntityID string = "ENTITY_TIMELINE"
const currentTickAttribute string = "CURRENT_TICK"
const actionStartTickAttribute string = "ACTION_START_TICK"


type Timeline struct {
	dao EntityDao
}

func NewTimeline(dao EntityDao) *Timeline {
	return &Timeline{
		dao: dao,
	}
}

func (t *Timeline) AddActions(actions []*Action) error {
	for _, action := range actions {
		err := t.AddAction(action)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Timeline) AddAction(action *Action) error {
	tick, err := t.GetCurrentTick()
	if err != nil {
		return err
	}
	return action.SetAttribute(actionStartTickAttribute, tick)
}

func (t *Timeline) RemoveAction(action *Action) error {
	return action.RemoveAttribute(actionStartTickAttribute)
}


func (t *Timeline) GetCurrentTick() (int64, error) {
	timelineEntities, err := t.dao.GetEntitiesWithAttributeType(currentTickAttribute)
	if err != nil {
		 return 0, err
	}
	if len(timelineEntities) > 1 {
		return 0, fmt.Errorf("expected 1 or 0 timelines but got %v", len(timelineEntities))
	}
	if len(timelineEntities) < 1 {
		return 0, fmt.Errorf("no timeline found, no current tick set")
	}
	currentTick, err := t.dao.GetAttribute(timelineEntities[0].GetID(),currentTickAttribute)
	if err != nil {
		return 0, err
	}
	if currentTick == nil {
		return 0, fmt.Errorf("no current tick")
	}
	return currentTick.(int64), err
}

func (t *Timeline) GetPendingAction() (*Action, error) {
	var earliestEndTick int64
	var earliestAction *Action
	tick, err := t.GetCurrentTick()
	if err != nil {
		return nil, err
	}
	actions, err := getActions(t.dao)
	if err != nil {
		return nil, err
	}
	for _, action := range actions {
		endTick, err := t.GetEndTickOfAction(action)
		if err != nil {
			return nil, err
		}
		if endTick <= tick && (earliestEndTick <= 0 || endTick < earliestEndTick) {
			earliestEndTick = endTick
			earliestAction = action
		}
	}
	return earliestAction, nil
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
	return startTick+duration, nil
}

func (t *Timeline) SetCurrentTick(tick int64) error {
	return t.dao.SetAttribute(timelineEntityID, currentTickAttribute, tick)
}


func indexOfAction(action *Action, actions []*Action) (int, error) {
	for index, a := range actions {
		if a == action {
			return index, nil
		}
	}
	return 0, fmt.Errorf("action %v not found in action array %v", action, actions)
}

func getActions(dao EntityDao) ([]*Action, error) {
	actionEntities, err := dao.GetEntitiesWithAttributeType(actionStartTickAttribute)
	if err != nil {
		return nil, err
	}
	actions := make([]*Action, len(actionEntities))
	for i, entity := range actionEntities {
		actions[i] = AsAction(entity)
	}
	return actions, nil
}

