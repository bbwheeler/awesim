package core_test

import (
	"errors"
	"testing"

	core "github.com/bbwheeler/awesim/core"
	"github.com/bbwheeler/awesim/core/storage"
)

func TestEndToEnd(t *testing.T) {
	const startTick core.Tick = core.Tick(1)

	entityStore := storage.NewEntityMapStore()
	actionStore := core.NewActionStore(entityStore)
	actorStore := core.NewActorStore(entityStore)
	timeline := core.NewTimeline(actionStore)
	actorOne, err := actorStore.NewActor()
	if err != nil {
		t.Error(err)
	}

	actionOne, err := actionStore.NewAction(actorOne.GetID(), 10)
	if err != nil {
		t.Error(err)
	}
	actorTwo, err := actorStore.NewActor()
	actionTwo, err := actionStore.NewAction(actorTwo.GetID(), 5)
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

	tick, err := timeline.GetStartTickOfAction(actionOne.GetID())
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
