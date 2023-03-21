package core

import "fmt"

const timelineEntity string = "ENTITY_TIMELINE"
const currentTick string = "CURRENT_TICK"
const actionStartTick string = "ACTION_START_TICK"


type Timeline struct {
	dao EntityDao
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
	return action.SetAttribute(actionStartTick, tick)
}

func (t *Timeline) RemoveAction(action *Action) error {
	return action.RemoveAttribute(actionStartTick)
}


func (t *Timeline) GetCurrentTick() (int64, error) {
	timelineEntities, err := t.dao.GetEntitiesWithAttributeType(currentTick)
	if err != nil {
		 return 0, err
	}
	if len(timelineEntities) > 1 {
		return 0, fmt.Errorf("expected 1 or 0 timelines but got %v", len(timelineEntities))
	}
	if len(timelineEntities) < 1 {
		timelineEntity := NewEntity(t.dao)
		timelineEntity.SetAttribute(currentTick, 0)
		timelineEntities = append(timelineEntities, timelineEntity)
	}
	currentTick, err := timelineEntities[0].GetAttribute(currentTick)
	return currentTick.(int64), err
}

func (t *Timeline) GetFirstPendingAction() (*Action, error) {
	var earliestStartTick int64
	var earliestAction *Action
	actions, err := getActions(t.dao)
	if err != nil {
		return nil, err
	}
	for _, action := range actions {
		startTick, err := t.GetStartTickOfAction(action)
		if err != nil {
			return nil, err
		}
		if earliestStartTick <= 0 || startTick < earliestStartTick {
			earliestStartTick = startTick
			earliestAction = action
		}
	}
	return earliestAction, nil
}

func (t *Timeline) GetStartTickOfAction(a *Action) (int64, error) {
	startTick, err := a.GetAttribute(actionStartTick)
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
	return t.dao.SetAttribute(timelineEntity, currentTick, tick)
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
	actionEntities, err := dao.GetEntitiesWithAttributeType(actionStartTick)
	if err != nil {
		return nil, err
	}
	actions := make([]*Action, len(actionEntities))
	for i, entity := range actionEntities {
		actions[i] = AsAction(entity)
	}
	return actions, nil
}

