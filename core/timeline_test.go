package core_test

import "testing"
import mapstore "github.com/bbwheeler/awesim/mapstore"
import core "github.com/bbwheeler/awesim/core"

func TestEndToEnd(t *testing.T) {
	const startTick int64 = int64(1)

	dao := mapstore.NewEntityDaoMapImpl()
	timeline := core.NewTimeline(dao)
	actorOne := core.NewActor(dao)
	actionOne, err := core.NewAction(actorOne, 10, dao)
	if err != nil {
		t.Error(err)
	}
	actorTwo := core.NewActor(dao)
	actionTwo, err := core.NewAction(actorTwo, 5, dao)
	if err != nil {
		t.Error(err)
	}

	err = timeline.SetCurrentTick(startTick)
	if err != nil {
		t.Fatal(err)
	}

	currentTick, err := timeline.GetCurrentTick()
	if err != nil {
		t.Fatal(err)
	}
	if startTick != currentTick {
		t.Fatalf("Expected tick to be %v but was %v", startTick, currentTick)
	}

	err = timeline.AddActions([]*core.Action{actionOne,actionTwo})
	if err != nil {
		t.Fatal(err)
	}

	tick, err := timeline.GetStartTickOfAction(actionOne)
	if err != nil {
		t.Fatal(err)
	}
	if tick != startTick {
		t.Fatalf("Expected action start tick to be %v but it was %v", startTick, tick)
	}

	nextAction, err := timeline.GetPendingAction()
	if err != nil {
		t.Fatal(err)
	}
	if nextAction != nil {
		t.Fatalf("Expected nil action but got %v\n", nextAction)
	}

	err = timeline.SetCurrentTick(startTick+7)
	if err != nil {
		t.Fatal(err)
	}

	nextAction, err = timeline.GetPendingAction()
	if err != nil {
		t.Fatal(err)
	}
	if nextAction == nil {
		t.Fatalf("Expected next action to be %v but was nil", actionTwo.GetID())
	}

	if nextAction.GetID() != actionTwo.GetID() {
		t.Fatalf("Expected action %v but got action %v", actionTwo.GetID(), nextAction.GetID())
	}
	
	err = timeline.RemoveAction(actionTwo)
	if err != nil {
		t.Fatal(err)
	}

	err = timeline.SetCurrentTick(startTick+11)
	if err != nil {
		t.Fatal(err)
	}

	nextAction, err = timeline.GetPendingAction()
	if err != nil {
		t.Fatal(err)
	}
	if nextAction.GetID() != actionOne.GetID() {
		t.Fatalf("Expected action %v but got action %v", actionOne.GetID(), nextAction.GetID())
	}


}

