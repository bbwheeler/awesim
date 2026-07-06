// engine_e2e_test.go
//
// NOTE: adjust the import path below ("github.com/bbwheeler/awesim/engine")
// if it differs from your actual module layout. core.Tick is assumed to be
// a plain integer-backed type (e.g. `type Tick int64`), which is the only
// assumption these fakes rely on.
package engine_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/bbwheeler/awesim/core"
	"github.com/bbwheeler/awesim/engine"
)

// ---------------------------------------------------------------------
// Fake entityProvider: simple in-memory entity/attribute store.
// ---------------------------------------------------------------------

type fakeEntityProvider struct {
	data    map[string]map[string]any
	removed []string
}

func newFakeEntityProvider() *fakeEntityProvider {
	return &fakeEntityProvider{data: map[string]map[string]any{}}
}

func (p *fakeEntityProvider) RemoveEntity(entityID string) error {
	delete(p.data, entityID)
	p.removed = append(p.removed, entityID)
	return nil
}

func (p *fakeEntityProvider) GetAttribute(entityId string, attributeId string) (any, error) {
	e, ok := p.data[entityId]
	if !ok {
		return nil, fmt.Errorf("entity %q not found", entityId)
	}
	v, ok := e[attributeId]
	if !ok {
		return nil, fmt.Errorf("attribute %q not found on entity %q", attributeId, entityId)
	}
	return v, nil
}

func (p *fakeEntityProvider) HasAttribute(entityId string, attributeId string) (bool, error) {
	e, ok := p.data[entityId]
	if !ok {
		return false, nil
	}
	_, ok = e[attributeId]
	return ok, nil
}

func (p *fakeEntityProvider) SetAttribute(entityId string, attributeId string, value any) error {
	e, ok := p.data[entityId]
	if !ok {
		e = map[string]any{}
		p.data[entityId] = e
	}
	e[attributeId] = value
	return nil
}

func (p *fakeEntityProvider) RemoveAttribute(entityId string, attributeId string) error {
	if e, ok := p.data[entityId]; ok {
		delete(e, attributeId)
	}
	return nil
}

func (p *fakeEntityProvider) GetEntitiesWithAttributes(attributes map[string]any) ([]string, error) {
	var ids []string
outer:
	for id, e := range p.data {
		for k, v := range attributes {
			if e[k] != v {
				continue outer
			}
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (p *fakeEntityProvider) GetEntitiesWithAttributeType(attribute string) ([]string, error) {
	var ids []string
	for id, e := range p.data {
		if _, ok := e[attribute]; ok {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (p *fakeEntityProvider) GetEntitiesWithAttribute(attribute string, value any) ([]string, error) {
	var ids []string
	for id, e := range p.data {
		if v, ok := e[attribute]; ok && v == value {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// ---------------------------------------------------------------------
// Fake timeline: tracks the current tick and a pending-action queue.
// Each pending action's tick is looked up via the shared entity store
// (attribute "tick"), the same way a real implementation would.
// ---------------------------------------------------------------------

type fakeTimeline struct {
	currentTick core.Tick
	entities    *fakeEntityProvider
	pending     []string
}

func (t *fakeTimeline) GetCurrentTick() (core.Tick, error) { return t.currentTick, nil }

func (t *fakeTimeline) SetCurrentTick(tick core.Tick) error {
	t.currentTick = tick
	return nil
}

func (t *fakeTimeline) GetPendingActionID() (string, error) {
	if len(t.pending) == 0 {
		return "", core.ErrNoPendingAction
	}
	return t.pending[0], nil
}

func (t *fakeTimeline) GetNextActionUpTo(maxTick core.Tick) (string, error) {
	bestIdx := -1
	var bestTick core.Tick
	for i, id := range t.pending {
		tickVal, err := t.entities.GetAttribute(id, "tick")
		if err != nil {
			continue
		}
		tick, ok := tickVal.(core.Tick)
		if !ok || tick > maxTick {
			continue
		}
		if bestIdx == -1 || tick < bestTick {
			bestIdx, bestTick = i, tick
		}
	}
	if bestIdx == -1 {
		return "", core.ErrNoPendingAction
	}
	id := t.pending[bestIdx]
	t.pending = append(t.pending[:bestIdx], t.pending[bestIdx+1:]...)
	return id, nil
}

func (t *fakeTimeline) AddActions(actionIDs []string) error {
	t.pending = append(t.pending, actionIDs...)
	return nil
}

// ---------------------------------------------------------------------
// Fake actor: hands out a scripted sequence of actions, one per tick.
// It reports "no pending action" once the entity for its current action
// has been removed (i.e. resolved), matching resolveAction's behavior.
// ---------------------------------------------------------------------

type plannedAction struct {
	id   string
	tick core.Tick
}

type fakeActor struct {
	id            string
	entities      *fakeEntityProvider
	currentAction string
	plan          []plannedAction
	planIndex     int
	provideCalls  int
	getNextErr    error
}

func (a *fakeActor) GetID() string { return a.id }

func (a *fakeActor) GetNextActionID() (string, error) {
	if a.getNextErr != nil {
		return "", a.getNextErr
	}
	if a.currentAction != "" {
		if _, err := a.entities.GetAttribute(a.currentAction, "tick"); err == nil {
			return a.currentAction, nil // still pending / not yet resolved
		}
		a.currentAction = "" // its entity is gone: it resolved
	}
	return "", nil
}

func (a *fakeActor) ProvideNextAction() (string, error) {
	a.provideCalls++
	if a.planIndex >= len(a.plan) {
		return "", nil // nothing left to do
	}
	next := a.plan[a.planIndex]
	a.planIndex++
	if err := a.entities.SetAttribute(next.id, "tick", next.tick); err != nil {
		return "", err
	}
	a.currentAction = next.id
	return next.id, nil
}

type fakeActorProvider struct {
	actors map[string]*fakeActor
	order  []string
}

func (p *fakeActorProvider) GetActor(actorID string) core.Actor {
	a, ok := p.actors[actorID]
	if !ok {
		return nil
	}
	return a
}

func (p *fakeActorProvider) GetAllActorIDs() ([]string, error) { return p.order, nil }

// ---------------------------------------------------------------------
// Fake action: on Resolve, bumps a counter on a shared "world" entity
// (unless configured to fail).
// ---------------------------------------------------------------------

type fakeAction struct {
	id         string
	entities   *fakeEntityProvider
	resolveErr error
	onResolve  func()
}

func (a *fakeAction) GetAttribute(attribute string) (any, error) {
	// TODO
	return nil, fmt.Errorf("unimplemented")
}
func (a *fakeAction) SetAttribute(attribute string, value any) error {
	// TODO
	return fmt.Errorf("unimplemented")
}
func (a *fakeAction) GetDuration() (core.Tick, error) {
	// TODO
	return 0, fmt.Errorf("unimplemented")
}

func (a *fakeAction) Resolve() error {
	if a.resolveErr != nil {
		return a.resolveErr
	}
	if a.onResolve != nil {
		a.onResolve()
	}
	current, _ := a.entities.GetAttribute("world", "eventsResolved")
	n, _ := current.(int)
	return a.entities.SetAttribute("world", "eventsResolved", n+1)
}

type fakeActionProvider struct {
	entities    *fakeEntityProvider
	resolveErrs map[string]error
	resolvedIDs []string
}

func (p *fakeActionProvider) GetAction(actionID string) core.Action {
	return &fakeAction{
		id:         actionID,
		entities:   p.entities,
		resolveErr: p.resolveErrs[actionID],
		onResolve: func() {
			p.resolvedIDs = append(p.resolvedIDs, actionID)
		},
	}
}

// ---------------------------------------------------------------------
// Assertions
// ---------------------------------------------------------------------

func assertEventsResolved(t *testing.T, entities *fakeEntityProvider, want int) {
	t.Helper()
	got, err := entities.GetAttribute("world", "eventsResolved")
	if err != nil {
		t.Fatalf("failed to read eventsResolved: %v", err)
	}
	if got.(int) != want {
		t.Fatalf("expected eventsResolved=%d, got %v", want, got)
	}
}

func assertEntityRemoved(t *testing.T, entities *fakeEntityProvider, id string) {
	t.Helper()
	if _, ok := entities.data[id]; ok {
		t.Fatalf("expected entity %q to have been removed", id)
	}
	for _, r := range entities.removed {
		if r == id {
			return
		}
	}
	t.Fatalf("expected RemoveEntity to have been called for %q", id)
}

// ---------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------

// This is the core end-to-end scenario: one actor working through a
// multi-tick plan via repeated RunCurrentTurn calls.
func TestEngine_EndToEnd_SingleActorMultipleTicks(t *testing.T) {
	entities := newFakeEntityProvider()
	entities.SetAttribute("world", "eventsResolved", 0)

	hero := &fakeActor{
		id:       "hero",
		entities: entities,
		plan: []plannedAction{
			{id: "action-0", tick: 0},
			{id: "action-1", tick: 1},
			{id: "action-2", tick: 2},
		},
	}

	actors := &fakeActorProvider{
		actors: map[string]*fakeActor{"hero": hero},
		order:  []string{"hero"},
	}
	actions := &fakeActionProvider{entities: entities}
	tl := &fakeTimeline{currentTick: 0, entities: entities}

	eng := engine.New(entities, actors, actions, tl)

	// Tick 0
	if err := eng.RunCurrentTurn(); err != nil {
		t.Fatalf("RunCurrentTurn() at tick 0: %v", err)
	}
	assertEventsResolved(t, entities, 1)
	assertEntityRemoved(t, entities, "action-0")

	// Tick 1
	if err := eng.RunCurrentTurn(); err != nil {
		t.Fatalf("RunCurrentTurn() at tick 1: %v", err)
	}
	assertEventsResolved(t, entities, 2)
	assertEntityRemoved(t, entities, "action-1")

	// Tick 2
	if err := eng.RunCurrentTurn(); err != nil {
		t.Fatalf("RunCurrentTurn() at tick 2: %v", err)
	}
	assertEventsResolved(t, entities, 3)
	assertEntityRemoved(t, entities, "action-2")

	// Tick 3: actor's plan is exhausted; engine just advances the tick.
	if err := eng.RunCurrentTurn(); err != nil {
		t.Fatalf("RunCurrentTurn() at tick 3: %v", err)
	}
	assertEventsResolved(t, entities, 3) // unchanged

	if tl.currentTick != 4 {
		t.Fatalf("expected currentTick=4, got %d", tl.currentTick)
	}
	if len(tl.pending) != 0 {
		t.Fatalf("expected no pending actions left, got %v", tl.pending)
	}
	// 5 calls: two during tick0 (action-0 resolves, then action-1 is
	// pre-queued), one each on tick1 and tick2, and one on tick3 where
	// the plan is found to be exhausted.
	if hero.provideCalls != 5 {
		t.Fatalf("expected 5 ProvideNextAction calls, got %d", hero.provideCalls)
	}
}

func TestEngine_MultipleActors_AllActionsResolvedAtSameTick(t *testing.T) {
	entities := newFakeEntityProvider()
	entities.SetAttribute("world", "eventsResolved", 0)

	alice := &fakeActor{id: "alice", entities: entities, plan: []plannedAction{{id: "alice-action", tick: 0}}}
	bob := &fakeActor{id: "bob", entities: entities, plan: []plannedAction{{id: "bob-action", tick: 0}}}

	actors := &fakeActorProvider{
		actors: map[string]*fakeActor{"alice": alice, "bob": bob},
		order:  []string{"alice", "bob"},
	}
	actions := &fakeActionProvider{entities: entities}
	tl := &fakeTimeline{currentTick: 0, entities: entities}

	eng := engine.New(entities, actors, actions, tl)

	if err := eng.ExecuteToTick(core.Tick(0)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertEventsResolved(t, entities, 2)
	assertEntityRemoved(t, entities, "alice-action")
	assertEntityRemoved(t, entities, "bob-action")
}

func TestEngine_ExecuteToTick_ActionBeyondTargetTickRemainsPending(t *testing.T) {
	entities := newFakeEntityProvider()
	hero := &fakeActor{id: "hero", entities: entities, plan: []plannedAction{{id: "action-future", tick: 5}}}
	actors := &fakeActorProvider{actors: map[string]*fakeActor{"hero": hero}, order: []string{"hero"}}
	actions := &fakeActionProvider{entities: entities}
	tl := &fakeTimeline{currentTick: 0, entities: entities}

	eng := engine.New(entities, actors, actions, tl)

	if err := eng.ExecuteToTick(core.Tick(2)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tl.pending) != 1 || tl.pending[0] != "action-future" {
		t.Fatalf("expected action-future to remain pending, got %v", tl.pending)
	}
	if _, ok := entities.data["action-future"]; !ok {
		t.Fatalf("expected action-future entity to still exist (not resolved)")
	}
}

func TestEngine_ExecuteToTick_CurrentTickAfterTargetTick_NoOp(t *testing.T) {
	entities := newFakeEntityProvider()
	hero := &fakeActor{id: "hero", entities: entities, plan: []plannedAction{{id: "action-x", tick: 0}}}
	actors := &fakeActorProvider{actors: map[string]*fakeActor{"hero": hero}, order: []string{"hero"}}
	actions := &fakeActionProvider{entities: entities}
	tl := &fakeTimeline{currentTick: 5, entities: entities}

	eng := engine.New(entities, actors, actions, tl)

	if err := eng.ExecuteToTick(core.Tick(2)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hero.provideCalls != 0 {
		t.Fatalf("actor should not be consulted when current tick is past target, got %d calls", hero.provideCalls)
	}
	if len(tl.pending) != 0 {
		t.Fatalf("expected no pending actions queued, got %v", tl.pending)
	}
}

func TestEngine_ExecuteToTick_NegativeCurrentTick_ReturnsError(t *testing.T) {
	entities := newFakeEntityProvider()
	actors := &fakeActorProvider{actors: map[string]*fakeActor{}}
	actions := &fakeActionProvider{entities: entities}
	tl := &fakeTimeline{currentTick: core.Tick(-1), entities: entities}

	eng := engine.New(entities, actors, actions, tl)

	if err := eng.ExecuteToTick(core.Tick(0)); err == nil {
		t.Fatal("expected an error for a negative current tick, got nil")
	}
}

func TestEngine_ExecuteToTick_ActorError_Propagates(t *testing.T) {
	entities := newFakeEntityProvider()
	boom := errors.New("actor exploded")
	hero := &fakeActor{id: "hero", entities: entities, getNextErr: boom}
	actors := &fakeActorProvider{actors: map[string]*fakeActor{"hero": hero}, order: []string{"hero"}}
	actions := &fakeActionProvider{entities: entities}
	tl := &fakeTimeline{currentTick: 0, entities: entities}

	eng := engine.New(entities, actors, actions, tl)

	err := eng.ExecuteToTick(core.Tick(0))
	if !errors.Is(err, boom) {
		t.Fatalf("expected actor error to propagate, got %v", err)
	}
}

func TestEngine_ExecuteToTick_ActionResolveError_PropagatesAndSkipsRemoval(t *testing.T) {
	entities := newFakeEntityProvider()
	hero := &fakeActor{id: "hero", entities: entities, plan: []plannedAction{{id: "bad-action", tick: 0}}}
	actors := &fakeActorProvider{actors: map[string]*fakeActor{"hero": hero}, order: []string{"hero"}}

	boom := errors.New("resolve exploded")
	actions := &fakeActionProvider{entities: entities, resolveErrs: map[string]error{"bad-action": boom}}
	tl := &fakeTimeline{currentTick: 0, entities: entities}

	eng := engine.New(entities, actors, actions, tl)

	err := eng.ExecuteToTick(core.Tick(0))
	if !errors.Is(err, boom) {
		t.Fatalf("expected resolve error to propagate, got %v", err)
	}
	if _, exists := entities.data["bad-action"]; !exists {
		t.Fatalf("expected bad-action entity to still exist since Resolve failed before removal")
	}
	if len(entities.removed) != 0 {
		t.Fatalf("expected RemoveEntity not to be called when Resolve fails, got %v", entities.removed)
	}
}
