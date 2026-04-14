package core_test

import (
	"errors"
	"testing"

	core "github.com/bbwheeler/awesim/core"
	mapstore "github.com/bbwheeler/awesim/mapstore"
)

func TestEndToEnd(t *testing.T) {
	const startTick int64 = int64(1)

	dao := mapstore.NewEntityMapStore()
	timeline := core.NewTimeline(dao)
	actorOne := core.NewActor(dao)
	actionOne, err := core.NewAction(actorOne.GetID(), 10, dao)
	if err != nil {
		t.Error(err)
	}
	actorTwo := core.NewActor(dao)
	actionTwo, err := core.NewAction(actorTwo.GetID(), 5, dao)
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

	err = timeline.AddActions([]string{actionOne.GetID(), actionTwo.GetID()})
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

	nextActionID, err := timeline.GetPendingActionID()
	if !errors.Is(err, core.ErrNoPendingAction) {
		t.Fatal(err)
	}
	if nextActionID != "" {
		t.Fatalf("Expected nil action but got %v\n", nextActionID)
	}

	err = timeline.SetCurrentTick(startTick + 7)
	if err != nil {
		t.Fatal(err)
	}

	nextActionID, err = timeline.GetPendingActionID()
	if err != nil {
		t.Fatal(err)
	}
	if nextActionID == "" {
		t.Fatalf("Expected next action to be %v but was nil", actionTwo.GetID())
	}

	if nextActionID != actionTwo.GetID() {
		t.Fatalf("Expected action %v but got action %v", actionTwo.GetID(), nextActionID)
	}

	err = timeline.RemoveAction(actionTwo)
	if err != nil {
		t.Fatal(err)
	}

	err = timeline.SetCurrentTick(startTick + 11)
	if err != nil {
		t.Fatal(err)
	}

	nextActionID, err = timeline.GetPendingActionID()
	if err != nil {
		t.Fatal(err)
	}
	if nextActionID != actionOne.GetID() {
		t.Fatalf("Expected action %v but got action %v", actionOne.GetID(), nextActionID)
	}

}
