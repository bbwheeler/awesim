package timeline

import (
	"github.com/bbwheeler/awesim/awesim_core/entity"
	"github.com/bbwheeler/awesim/awesim_core/action"
)

const ACTION_ATTRIBUTE_TICK string = "ACTION TICK"

type timeline struct {
	dao entity.EntityDao
}

func (t timeline) addActionAtTick(actionId string, tick int64) {
	t.dao.SetIntegerAttribute(actionId, ACTION_ATTRIBUTE_TICK, tick)
}

func (t timeline) RemoveActionFromTimeline(actionId string) {
	t.dao.RemoveIntegerAttribute(actionId, ACTION_ATTRIBUTE_TICK)
}

func (t timeline) getFirstAction() string {
	var allActions []string = t.getAllActions()
	var firstAction string
	var minTick int64 = -1
	for _, actionId := range allActions {
		var tick = t.dao.GetIntegerAttribute(actionId, ACTION_ATTRIBUTE_TICK)
		if (minTick < 0 || tick < minTick) {
			firstAction = actionId
			minTick = tick
		}
	}
	return firstAction
}

func (t timeline) getAllActions() []string {
	return t.dao.GetEntitiesWithStringAttribute(entity.TypeAttribute, action.ActionTypeAttributeValue)
}

func (t timeline) getTickOfAction(actionId string) int64 {
	return t.dao.GetIntegerAttribute(actionId, ACTION_ATTRIBUTE_TICK)
}

func (t timeline) getTickOfNextAction() int64 {
	var actionId string = t.getFirstAction()
	if actionId != "" {
		return t.getTickOfAction(actionId)
	}
	return -1;
}

func (t timeline) clearAllActions() {
	for _, actionId := range t.getAllActions() {
		t.RemoveActionFromTimeline(actionId)
	}
}

func (t timeline) getAllActionsAtTick(tick int64) []string {
	return t.dao.GetEntitiesWithIntegerAttribute(ACTION_ATTRIBUTE_TICK, tick)	
}
