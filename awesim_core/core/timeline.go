package core

import "fmt"

type Timeline struct {
	actions []Action
}

func (t *Timeline) GetFirstPendingAction() (Action, error) {
	var earliestStartTick int64
	var earliestAction Action
	for _, action := range t.actions {
		startTick, err := action.GetStartTick()
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


func (t *Timeline) AddActions(newActions []Action) {
	t.actions = append(t.actions, newActions...)
}
func (t *Timeline) RemoveAction(action Action) error {
	index, err := indexOfAction(action, t.actions)
	if err != nil {
		return err
	}
	t.actions[index] = t.actions[len(t.actions)-1]
	t.actions = t.actions[:len(t.actions)-1]
	return nil
}

func indexOfAction(action Action, actions []Action) (int, error) {
	for index, a := range actions {
		if a == action {
			return index, nil
		}
	}
	return 0, fmt.Errorf("action %v not found in action array %v", action, actions)
}