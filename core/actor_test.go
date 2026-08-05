package core_test

// NOTE: This file assumes the core package's import path is
// "github.com/bbwheeler/awesim/core". Adjust the import below if that's wrong.

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bbwheeler/awesim/core"
)

// actionInvoker is an unexported constant inside core (value: "ACTION_INVOKER",
// per the package author). A couple of assertions below hardcode that literal
// to confirm GetNextActionID queries by the correct attribute -- if that
// internal constant's value ever changes, only those specific assertions need
// updating, not the rest of the suite.
const actionInvokerAttr = "ACTION_INVOKER"

// ---------------------------------------------------------------------------
// Fake EntityStore
// ---------------------------------------------------------------------------

type setAttributeCall struct {
	entityId    string
	attributeId string
	value       interface{}
}

type getEntitiesWithAttributeCall struct {
	attributeID    string
	attributeValue any
}

type fakeEntityStore struct {
	newEntityFunc  func() (*core.Entity, error)
	newEntityCalls int

	// Unused by ActorStore's current logic, but required to satisfy the
	// EntityStore interface. Left nil; will panic loudly if ever invoked,
	// which signals a test needs updating.
	getAttributeFunc                 func(entityId, attributeId string) (interface{}, error)
	hasAttributeFunc                 func(entityId, attributeId string) (bool, error)
	removeAttributeFunc              func(entityId, attributeId string) error
	getEntitiesWithAttributeTypeFunc func(attribute string) ([]string, error)

	setAttributeFunc  func(entityId, attributeId string, value interface{}) error
	setAttributeCalls []setAttributeCall

	getEntitiesWithAttributeFunc  func(attributeID string, attributeValue any) ([]string, error)
	getEntitiesWithAttributeCalls []getEntitiesWithAttributeCall
}

func (f *fakeEntityStore) NewEntity() (*core.Entity, error) {
	f.newEntityCalls++
	return f.newEntityFunc()
}
func (f *fakeEntityStore) GetAttribute(entityId, attributeId string) (interface{}, error) {
	return f.getAttributeFunc(entityId, attributeId)
}
func (f *fakeEntityStore) HasAttribute(entityId, attributeId string) (bool, error) {
	return f.hasAttributeFunc(entityId, attributeId)
}
func (f *fakeEntityStore) SetAttribute(entityId, attributeId string, value interface{}) error {
	f.setAttributeCalls = append(f.setAttributeCalls, setAttributeCall{entityId, attributeId, value})
	return f.setAttributeFunc(entityId, attributeId, value)
}
func (f *fakeEntityStore) RemoveAttribute(entityId, attributeId string) error {
	return f.removeAttributeFunc(entityId, attributeId)
}
func (f *fakeEntityStore) GetEntitiesWithAttribute(attributeID string, attributeValue any) ([]string, error) {
	f.getEntitiesWithAttributeCalls = append(f.getEntitiesWithAttributeCalls, getEntitiesWithAttributeCall{attributeID, attributeValue})
	return f.getEntitiesWithAttributeFunc(attributeID, attributeValue)
}
func (f *fakeEntityStore) GetEntitiesWithAttributeType(attribute string) ([]string, error) {
	return f.getEntitiesWithAttributeTypeFunc(attribute)
}

// ---------------------------------------------------------------------------
// NewActorStore
// ---------------------------------------------------------------------------

func TestNewActorStore(t *testing.T) {
	fes := &fakeEntityStore{}

	as := core.NewActorStore(fes)

	if as == nil {
		t.Fatal("expected NewActorStore to return a non-nil ActorStore")
	}
	if as.EntityStore != fes {
		t.Error("expected ActorStore to wrap the provided EntityStore")
	}
}

// ---------------------------------------------------------------------------
// NewActor
// ---------------------------------------------------------------------------

func TestActorStore_NewActor_Success(t *testing.T) {
	fes := &fakeEntityStore{}
	fes.newEntityFunc = func() (*core.Entity, error) {
		return core.EntityFromID("entity-1", fes), nil
	}
	fes.setAttributeFunc = func(entityId, attributeId string, value interface{}) error {
		return nil
	}

	as := core.NewActorStore(fes)

	entity, err := as.NewActor()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if entity == nil {
		t.Fatal("expected a non-nil entity")
	}
	if entity.GetID() != "entity-1" {
		t.Errorf("expected entity ID 'entity-1', got %q", entity.GetID())
	}

	if len(fes.setAttributeCalls) != 1 {
		t.Fatalf("expected 1 SetAttribute call, got %d", len(fes.setAttributeCalls))
	}
	call := fes.setAttributeCalls[0]
	if call.entityId != "entity-1" || call.attributeId != core.IsActorAttribute || call.value != true {
		t.Errorf("expected SetAttribute(entity-1, %q, true), got SetAttribute(%q, %q, %v)",
			core.IsActorAttribute, call.entityId, call.attributeId, call.value)
	}
}

func TestActorStore_NewActor_NewEntityError(t *testing.T) {
	wantErr := errors.New("new entity failed")
	fes := &fakeEntityStore{
		newEntityFunc: func() (*core.Entity, error) { return nil, wantErr },
	}

	as := core.NewActorStore(fes)

	entity, err := as.NewActor()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if entity != nil {
		t.Errorf("expected nil entity on error, got %v", entity)
	}
	if len(fes.setAttributeCalls) != 0 {
		t.Errorf("SetAttribute should not have been called, got %v", fes.setAttributeCalls)
	}
}

func TestActorStore_NewActor_SetAttributeError(t *testing.T) {
	wantErr := errors.New("set attribute failed")
	fes := &fakeEntityStore{}
	fes.newEntityFunc = func() (*core.Entity, error) {
		return core.EntityFromID("entity-1", fes), nil
	}
	fes.setAttributeFunc = func(entityId, attributeId string, value interface{}) error {
		return wantErr
	}

	as := core.NewActorStore(fes)

	entity, err := as.NewActor()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if entity != nil {
		t.Errorf("expected nil entity on error, got %v", entity)
	}
}

// ---------------------------------------------------------------------------
// GetAllActorIDs
// ---------------------------------------------------------------------------

func TestActorStore_GetAllActorIDs_Success(t *testing.T) {
	want := []string{"actor-1", "actor-2"}
	fes := &fakeEntityStore{
		getEntitiesWithAttributeFunc: func(attributeID string, attributeValue any) ([]string, error) {
			return want, nil
		},
	}

	as := core.NewActorStore(fes)

	got, err := as.GetAllActorIDs()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}

	if len(fes.getEntitiesWithAttributeCalls) != 1 {
		t.Fatalf("expected 1 call to GetEntitiesWithAttribute, got %d", len(fes.getEntitiesWithAttributeCalls))
	}
	call := fes.getEntitiesWithAttributeCalls[0]
	if call.attributeID != core.IsActorAttribute || call.attributeValue != true {
		t.Errorf("expected GetEntitiesWithAttribute(%q, true), got (%q, %v)",
			core.IsActorAttribute, call.attributeID, call.attributeValue)
	}
}

func TestActorStore_GetAllActorIDs_Error(t *testing.T) {
	wantErr := errors.New("lookup failed")
	fes := &fakeEntityStore{
		getEntitiesWithAttributeFunc: func(string, any) ([]string, error) { return nil, wantErr },
	}

	as := core.NewActorStore(fes)

	_, err := as.GetAllActorIDs()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

// ---------------------------------------------------------------------------
// GetNextActionID
// ---------------------------------------------------------------------------

func TestActorStore_GetNextActionID_Success(t *testing.T) {
	fes := &fakeEntityStore{
		getEntitiesWithAttributeFunc: func(attributeID string, attributeValue any) ([]string, error) {
			return []string{"action-1"}, nil
		},
	}

	as := core.NewActorStore(fes)

	got, err := as.GetNextActionID("actor-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != "action-1" {
		t.Errorf("expected 'action-1', got %q", got)
	}

	if len(fes.getEntitiesWithAttributeCalls) != 1 {
		t.Fatalf("expected 1 call to GetEntitiesWithAttribute, got %d", len(fes.getEntitiesWithAttributeCalls))
	}
	call := fes.getEntitiesWithAttributeCalls[0]
	if call.attributeID != actionInvokerAttr || call.attributeValue != "actor-1" {
		t.Errorf("expected GetEntitiesWithAttribute(%q, actor-1), got (%q, %v)",
			actionInvokerAttr, call.attributeID, call.attributeValue)
	}
}

func TestActorStore_GetNextActionID_NoneFound(t *testing.T) {
	fes := &fakeEntityStore{
		getEntitiesWithAttributeFunc: func(string, any) ([]string, error) { return nil, nil },
	}

	as := core.NewActorStore(fes)

	got, err := as.GetNextActionID("actor-1")
	if !errors.Is(err, core.ErrActionNotFoundForActor) {
		t.Fatalf("expected ErrActionNotFoundForActor, got %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestActorStore_GetNextActionID_MultipleFound(t *testing.T) {
	fes := &fakeEntityStore{
		getEntitiesWithAttributeFunc: func(string, any) ([]string, error) {
			return []string{"action-1", "action-2"}, nil
		},
	}

	as := core.NewActorStore(fes)

	got, err := as.GetNextActionID("actor-1")
	if err == nil {
		t.Fatal("expected an error when multiple actions are found, got nil")
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
	if errors.Is(err, core.ErrActionNotFoundForActor) {
		t.Error("expected a distinct 'too many actions found' error, not ErrActionNotFoundForActor")
	}
}

func TestActorStore_GetNextActionID_StoreError(t *testing.T) {
	wantErr := errors.New("lookup failed")
	fes := &fakeEntityStore{
		getEntitiesWithAttributeFunc: func(string, any) ([]string, error) { return nil, wantErr },
	}

	as := core.NewActorStore(fes)

	got, err := as.GetNextActionID("actor-1")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
