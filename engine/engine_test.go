package engine_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/bbwheeler/awesim/core"
	"github.com/bbwheeler/awesim/engine"
)

type fakeEntityStore struct {
	removeEntityCalls []string
	removeEntityErr   error
}

func (f *fakeEntityStore) RemoveEntity(entityID string) error {
	f.removeEntityCalls = append(f.removeEntityCalls, entityID)
	return f.removeEntityErr
}

type fakeActorStore struct {
	getAllActorIDsFunc   func() ([]string, error)
	getAllActorIDsCalls  int
	getNextActionIDFunc  func(actorID string) (string, error)
	getNextActionIDCalls []string
}

func (f *fakeActorStore) GetAllActorIDs() ([]string, error) {
	f.getAllActorIDsCalls++
	return f.getAllActorIDsFunc()
}
func (f *fakeActorStore) GetNextActionID(actorID string) (string, error) {
	f.getNextActionIDCalls = append(f.getNextActionIDCalls, actorID)
	return f.getNextActionIDFunc(actorID)
}

type fakeTimeline struct {
	getCurrentTickFunc func() (core.Tick, error)

	setCurrentTickFunc  func(tick core.Tick) error
	setCurrentTickCalls []core.Tick

	getPendingActionIDFunc func() (string, error)

	getNextActionUpToFunc  func(maxTick core.Tick) (string, error)
	getNextActionUpToCalls []core.Tick

	addActionsFunc  func(actionIDs []string) error
	addActionsCalls [][]string
}

func (f *fakeTimeline) GetCurrentTick() (core.Tick, error) {
	return f.getCurrentTickFunc()
}
func (f *fakeTimeline) SetCurrentTick(tick core.Tick) error {
	f.setCurrentTickCalls = append(f.setCurrentTickCalls, tick)
	return f.setCurrentTickFunc(tick)
}
func (f *fakeTimeline) GetPendingActionID() (string, error) {
	return f.getPendingActionIDFunc()
}
func (f *fakeTimeline) GetNextActionUpTo(maxTick core.Tick) (string, error) {
	f.getNextActionUpToCalls = append(f.getNextActionUpToCalls, maxTick)
	return f.getNextActionUpToFunc(maxTick)
}
func (f *fakeTimeline) AddActions(actionIDs []string) error {
	f.addActionsCalls = append(f.addActionsCalls, actionIDs)
	return f.addActionsFunc(actionIDs)
}

type fakeActionProvider struct {
	provideNextActionForFunc func(actorID string) (string, error)
	calls                    []string
}

func (f *fakeActionProvider) ProvideNextActionFor(actorID string) (string, error) {
	f.calls = append(f.calls, actorID)
	return f.provideNextActionForFunc(actorID)
}

type fakeActionResolver struct {
	resolveActionFunc func(actionID string) error
	calls             []string
}

func (f *fakeActionResolver) ResolveAction(actionID string) error {
	f.calls = append(f.calls, actionID)
	return f.resolveActionFunc(actionID)
}

func TestNew_ReturnsFunctioningEngine(t *testing.T) {
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(0), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			return "", core.ErrNoPendingAction
		},
		addActionsFunc:     func([]string) error { return nil },
		setCurrentTickFunc: func(core.Tick) error { return nil },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if e == nil {
		t.Fatal("expected New to return a non-nil Engine")
	}
	if err := e.RunCurrentTurn(); err != nil {
		t.Fatalf("expected wired engine to run a turn without error, got %v", err)
	}
}

func TestRunCurrentTurn_Success(t *testing.T) {
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			return "", core.ErrNoPendingAction
		},
		addActionsFunc:     func([]string) error { return nil },
		setCurrentTickFunc: func(core.Tick) error { return nil },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if err := e.RunCurrentTurn(); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(ts.setCurrentTickCalls) != 1 || ts.setCurrentTickCalls[0] != core.Tick(6) {
		t.Errorf("expected SetCurrentTick(6), got calls %v", ts.setCurrentTickCalls)
	}
}

func TestRunCurrentTurn_GetCurrentTickError(t *testing.T) {
	wantErr := errors.New("boom")
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(0), wantErr },
	}
	e := engine.New(&fakeEntityStore{}, &fakeActorStore{}, ts, &fakeActionProvider{}, &fakeActionResolver{})

	err := e.RunCurrentTurn()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if len(ts.setCurrentTickCalls) != 0 {
		t.Errorf("SetCurrentTick should not have been called, got %v", ts.setCurrentTickCalls)
	}
}

func TestRunCurrentTurn_NegativeTick(t *testing.T) {
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(-1), nil },
	}
	e := engine.New(&fakeEntityStore{}, &fakeActorStore{}, ts, &fakeActionProvider{}, &fakeActionResolver{})

	err := e.RunCurrentTurn()
	if err == nil {
		t.Fatal("expected an error for negative tick, got nil")
	}
	if len(ts.setCurrentTickCalls) != 0 {
		t.Errorf("SetCurrentTick should not have been called, got %v", ts.setCurrentTickCalls)
	}
}

func TestRunCurrentTurn_ExecuteToTickError(t *testing.T) {
	wantErr := errors.New("actor lookup failed")
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, wantErr },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(1), nil },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	err := e.RunCurrentTurn()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if len(ts.setCurrentTickCalls) != 0 {
		t.Errorf("SetCurrentTick should not have been called, got %v", ts.setCurrentTickCalls)
	}
}

func TestRunCurrentTurn_SetCurrentTickError(t *testing.T) {
	wantErr := errors.New("persist failed")
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(3), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			return "", core.ErrNoPendingAction
		},
		addActionsFunc:     func([]string) error { return nil },
		setCurrentTickFunc: func(core.Tick) error { return wantErr },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	err := e.RunCurrentTurn()
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestExecuteToTick_GetCurrentTickError(t *testing.T) {
	wantErr := errors.New("boom")
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(0), wantErr },
	}
	e := engine.New(&fakeEntityStore{}, &fakeActorStore{}, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(10)); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestExecuteToTick_NegativeCurrentTick(t *testing.T) {
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(-5), nil },
	}
	e := engine.New(&fakeEntityStore{}, &fakeActorStore{}, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(10)); err == nil {
		t.Fatal("expected an error for negative current tick, got nil")
	}
}

func TestExecuteToTick_CurrentGreaterThanTarget(t *testing.T) {
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(10), nil },
		// getNextActionUpToFunc and addActionsFunc intentionally left nil:
		// if the loop body runs at all, calling them will panic and fail
		// the test loudly.
	}
	as := &fakeActorStore{
		// getAllActorIDsFunc intentionally nil for the same reason.
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(5)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if as.getAllActorIDsCalls != 0 {
		t.Errorf("expected no actor lookups when current > target, got %d calls", as.getAllActorIDsCalls)
	}
}

func TestExecuteToTick_NoPendingActions(t *testing.T) {
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			return "", core.ErrNoPendingAction
		},
		addActionsFunc: func([]string) error { return nil },
	}
	ar := &fakeActionResolver{}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, ar)

	if err := e.ExecuteToTick(core.Tick(5)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ar.calls) != 0 {
		t.Errorf("expected no actions resolved, got %v", ar.calls)
	}
}

func TestExecuteToTick_ResolvesMultipleActionsUntilDrained(t *testing.T) {
	actionSequence := []string{"action-1", "action-2", "action-3"}
	callIndex := 0

	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			if callIndex >= len(actionSequence) {
				return "", core.ErrNoPendingAction
			}
			id := actionSequence[callIndex]
			callIndex++
			return id, nil
		},
		addActionsFunc: func([]string) error { return nil },
	}
	ar := &fakeActionResolver{
		resolveActionFunc: func(actionID string) error { return nil },
	}
	es := &fakeEntityStore{}
	e := engine.New(es, as, ts, &fakeActionProvider{}, ar)

	if err := e.ExecuteToTick(core.Tick(5)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !reflect.DeepEqual(ar.calls, actionSequence) {
		t.Errorf("expected resolver calls %v, got %v", actionSequence, ar.calls)
	}
	if !reflect.DeepEqual(es.removeEntityCalls, actionSequence) {
		t.Errorf("expected RemoveEntity calls %v, got %v", actionSequence, es.removeEntityCalls)
	}
	// The action-provisioning pass runs once per loop iteration: 3
	// successful resolutions + 1 final pass that discovers
	// ErrNoPendingAction = 4 passes total.
	if as.getAllActorIDsCalls != 4 {
		t.Errorf("expected 4 action-provisioning passes, got %d", as.getAllActorIDsCalls)
	}
}

func TestExecuteToTick_ActorLookupError(t *testing.T) {
	wantErr := errors.New("actor store failed")
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, wantErr },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(5)); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

// This exercises the "stop processing actors as soon as one lookup errors"
// behavior. It's driven entirely through ExecuteToTick: when GetNextActionID
// errors for the second of three actors, the third actor should never be
// queried, and the error should propagate out of ExecuteToTick.
func TestExecuteToTick_ActorNeedsActionLookupErrorStopsEarly(t *testing.T) {
	wantErr := errors.New("actor lookup failed")
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) {
			return []string{"a1", "a2", "a3"}, nil
		},
		getNextActionIDFunc: func(actorID string) (string, error) {
			if actorID == "a2" {
				return "", wantErr
			}
			return "", nil
		},
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(5)); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if len(as.getNextActionIDCalls) != 2 {
		t.Errorf("expected processing to stop after a2 (2 calls), got %v", as.getNextActionIDCalls)
	}
}

func TestExecuteToTick_ProvideNextActionForError(t *testing.T) {
	wantErr := errors.New("provider failed")
	as := &fakeActorStore{
		getAllActorIDsFunc:  func() ([]string, error) { return []string{"a1"}, nil },
		getNextActionIDFunc: func(string) (string, error) { return "", nil },
	}
	ap := &fakeActionProvider{
		provideNextActionForFunc: func(string) (string, error) { return "", wantErr },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		// addActionsFunc intentionally nil: must not be reached.
	}
	e := engine.New(&fakeEntityStore{}, as, ts, ap, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(5)); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if len(ts.addActionsCalls) != 0 {
		t.Errorf("AddActions should not have been called, got %v", ts.addActionsCalls)
	}
}

func TestExecuteToTick_EmptyProvidedActionIsSkipped(t *testing.T) {
	as := &fakeActorStore{
		getAllActorIDsFunc:  func() ([]string, error) { return []string{"a1", "a2"}, nil },
		getNextActionIDFunc: func(string) (string, error) { return "", nil },
	}
	ap := &fakeActionProvider{
		provideNextActionForFunc: func(actorID string) (string, error) {
			if actorID == "a1" {
				return "new-action-1", nil
			}
			return "", nil // a2 produces no action
		},
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			return "", core.ErrNoPendingAction
		},
		addActionsFunc: func([]string) error { return nil },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, ap, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(5)); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(ts.addActionsCalls) != 1 || !reflect.DeepEqual(ts.addActionsCalls[0], []string{"new-action-1"}) {
		t.Errorf("expected AddActions([new-action-1]), got %v", ts.addActionsCalls)
	}
}

func TestExecuteToTick_AddActionsError(t *testing.T) {
	wantErr := errors.New("add actions failed")
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		addActionsFunc:     func([]string) error { return wantErr },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(5)); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestExecuteToTick_GetNextActionUpToError(t *testing.T) {
	wantErr := errors.New("timeline read failed")
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			return "", wantErr
		},
		addActionsFunc: func([]string) error { return nil },
	}
	e := engine.New(&fakeEntityStore{}, as, ts, &fakeActionProvider{}, &fakeActionResolver{})

	if err := e.ExecuteToTick(core.Tick(5)); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func TestExecuteToTick_ResolveActionError(t *testing.T) {
	wantErr := errors.New("resolve failed")
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			return "action-1", nil
		},
		addActionsFunc: func([]string) error { return nil },
	}
	ar := &fakeActionResolver{
		resolveActionFunc: func(string) error { return wantErr },
	}
	es := &fakeEntityStore{}
	e := engine.New(es, as, ts, &fakeActionProvider{}, ar)

	if err := e.ExecuteToTick(core.Tick(5)); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
	if len(es.removeEntityCalls) != 0 {
		t.Errorf("RemoveEntity should not be called when ResolveAction fails, got %v", es.removeEntityCalls)
	}
}

func TestExecuteToTick_RemoveEntityError(t *testing.T) {
	wantErr := errors.New("remove entity failed")
	as := &fakeActorStore{
		getAllActorIDsFunc: func() ([]string, error) { return nil, nil },
	}
	ts := &fakeTimeline{
		getCurrentTickFunc: func() (core.Tick, error) { return core.Tick(5), nil },
		getNextActionUpToFunc: func(core.Tick) (string, error) {
			return "action-1", nil
		},
		addActionsFunc: func([]string) error { return nil },
	}
	ar := &fakeActionResolver{
		resolveActionFunc: func(string) error { return nil },
	}
	es := &fakeEntityStore{
		removeEntityErr: wantErr,
	}
	e := engine.New(es, as, ts, &fakeActionProvider{}, ar)

	if err := e.ExecuteToTick(core.Tick(5)); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}
