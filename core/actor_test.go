package core_test

import (
	"testing"

	core "github.com/bbwheeler/awesim/core"
	"github.com/bbwheeler/awesim/core/storage"
)

func TestNewActor(t *testing.T) {

	dao := storage.NewEntityMapStore()

	testActor := core.NewActor(dao)

	isActor, err := dao.GetAttribute(testActor.GetID(), core.IsActorAttribute)

	if err != nil {
		t.Fatal(err)
	}
	if !isActor.(bool) {
		t.Fatalf("Actor %v is not an actor\n", testActor)
	}
}
