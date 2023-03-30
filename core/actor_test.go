package core_test

import "testing"
import mapstore "github.com/bbwheeler/awesim/mapstore"
import core "github.com/bbwheeler/awesim/core"

func TestNewActor(t *testing.T) {

	dao := mapstore.NewEntityDaoMapImpl()

	testActor := core.NewActor(dao)

	isActor, err := testActor.GetAttribute(core.IsActorAttribute)

	if err != nil {
		t.Fatal(err)
	}
	if !isActor.(bool) {
		t.Fatalf("Actor %v is not an actor\n", testActor)
	}
}

