# AGENTS.md

## Architecture

**awesim** is a Go ECS (Entity-Component-System) game simulation engine. It models turn-based simulation via an Entity–Action–Timeline loop. **Every type and method in the repo is exported/uppercase.** ARCHITECTURE.md's "deliberately sealed / lowercase internal" claim is wrong and describes a different (planned) architecture, not this codebase.

### Package boundaries

| Package | Purpose |
|---------|---------|
| `core` | Public types `Entity`, `Actor`, `Action`, `Timeline`; public interface `EntityStore`; public helper constructors `NewEntity`, `GetEntity`, `NewActor`, `GetActor`, `GetAllActors(EntityProvider)`, `NewAction`, `GetAction`, `NewTimeline`. |
| `core/storage` | `EntityMapStore` (public struct) + `NewEntityMapStore()`. Single backing flat `map[string]any` keyed by `entityID#attributeID`. |
| `core/action_deciders` | `DoNothing` (public) + `NewDoNothingDecider(entityStore, timeline)`, exported constant `ActionTypeDoNothing = "DoNothing"`. |
| `core/action_resolvers` | Exported `ActionResolver` interface (`ResolveAction(string) (bool, error)`), `NoAction` struct, `Ordered` composite (`NewOrdered([]ActionResolver)`). |
| `engine` | `Engine` struct + `New(entityProvider, actionDecider, timeline, actionResolver)` (all four are unexported structural interfaces defined *inside* the `engine` package, not the `core` ones). Methods: `Run() error`, `RunCurrentTurn() error`, `ExecuteToTick(int64) error`. |
| `example` | Demo wiring. Not a library. |

### Exported vs unexported inside the `core` package

Exported from `core`: `EntityStore`, `Entity`, `Actor`, `Action`, `Timeline`, `Attribute` (generic constraint), `EntityProvider`, `IsActorAttribute` (`"IsActor"`), `ErrActionNotFoundForActor`, `ErrNoPendingAction`.

Unexported attribute-key constants (used by the engine and `core` internally, not stable public API):
```
actionInvoker              = "ACTION_INVOKER"
actionDuration             = "ACTION_DURATION"
timelineEntityID           = "ENTITY_TIMELINE"
currentTickAttribute       = "CURRENT_TICK"
actionStartTickAttribute   = "ACTION_START_TICK"
```

### Entity model (actually, from source)

- Entities have no type or state — they are just `(attributeID, value)` pairs keyed in the store. Two wrappers (`Actor`, `Action`) embed `Entity` and differ only by helper methods.
- **Actor-ness is a sentinel attribute**: `NewActor` sets `IsActor=true`. `GetAllActors(provider)` calls `provider.GetEntitiesWithAttribute("IsActor", true)`.
- **An Action is an entity with** `ACTION_INVOKER=<actorID>` and `ACTION_DURATION=<int64>`. When scheduled (added to a timeline), the timeline also writes `ACTION_START_TICK=<currentTick>` onto the same entity.
- **The timeline is not itself an entity.** `core.NewTimeline(store)` returns a `*Timeline` struct wrapping the store. Tick state lives somewhere in the store — whichever entity currently has a `CURRENT_TICK` attribute. There is one hardcoded entity ID used for **writing** the tick: `"ENTITY_TIMELINE"`. **Reading** the tick is generic: `Timeline.GetCurrentTick()` calls `store.GetEntitiesWithAttributeType("CURRENT_TICK")` and returns whichever entity (actor, custom, whatever) currently carries it. See "Gotchas" for the interaction between these two.

### Interface wiring (what's actually passed around)

- `engine.New` takes **four** unexported interfaces: `entityProvider`, `actionDecider`, `timeline`, `actionResolver`. These are **structural** — any value satisfying them works. `*storage.EntityMapStore` satisfies `entityProvider`, `*core.Timeline` satisfies `timeline`, `*action_deciders.DoNothing` satisfies `actionDecider`, and `*action_resolvers.NoAction` / `*action_resolvers.Ordered` both satisfy `actionResolver`. That is exactly how `example/main.go` wires it, and it compiles and runs.
- `engine.Actor` (a second, different interface: `GetNextActionID()` + `GetID()`) and `engine.actorProvider` have **no implementations anywhere in the repo.** `core.Actor.GetNextAction()` does not match (`GetNextAction`, not `GetNextActionID`). These are dead types.
- `core/action_resolvers.ActionResolver` (exported) and `engine.actionResolver` (unexported) have **identical signatures** but are distinct Go types. `Ordered` and `NoAction` satisfy both because Go interfaces are structural.
- `core.Attribute` is a generic type constraint (`string | int64 | float64 | bool`) declared in `core/entity.go:11` and **referenced nowhere else in the codebase.** It is not actually used.

### Module wiring (verified from `example/main.go` + `engine/Engine.go`)

```go
store        := storage.NewEntityMapStore()            // satisfies engine.entityProvider
timeline     := core.NewTimeline(store)                // satisfies engine.timeline
actionDecider:= action_deciders.NewDoNothingDecider(store, timeline) // satisfies engine.actionDecider
actionResolver := &action_resolvers.NoAction{}         // satisfies engine.actionResolver  (ALSO: action_resolvers.NewOrdered([]ActionResolver{...}))
eng          := engine.New(store, actionDecider, timeline, actionResolver)

timeline.SetCurrentTick(1)   // REQUIRED before any engine call — writes CURRENT_TICK onto entityID "ENTITY_TIMELINE"
err := eng.RunCurrentTurn()  // one tick
err = eng.Run()              // infinite loop (see below) or until it errors
```

### Engine execution model (from `engine/Engine.go`, verified)

- `Run()` is `for !ended { RunCurrentTurn() }` where `ended` is a local `var ended bool` that is **never set.** `Run()` is either an infinite loop or terminates on the first error from `RunCurrentTurn`.
- `RunCurrentTurn()` = `GetCurrentTick()` → `ExecuteToTick(currentTick)` → `SetCurrentTick(currentTick+1)`.
- `ExecuteToTick(targetTick)` is `for { decideActions(); id = GetPendingActionIDWithMaxTick(targetTick); if ErrNoPendingAction → return nil; resolveAction(id) }`.
- `decideActions()` = `getActorsThatNeedActions()` → for each actor `actionDecider.DecideActionForActor(actorID)` → `timeline.AddActions(newActionIDs)`.
- `getActorsThatNeedActions()` = `core.GetAllActors(entityProvider)` → for each actor, `actor.GetNextAction()`; appends to result **only if `err == nil && actionID == ""`**.
- `resolveAction(id)` = `actionResolver.ResolveAction(id)`; if `(false, _)` → `return fmt.Errorf("action %v was not resolved", id)`; if `(true, _)` → `entityProvider.RemoveEntity(id)`.

## Commands

```bash
go mod tidy            # add/update dependencies
go test ./...          # all tests pass; coverage is partial (see below)
go run example/main.go # prints "Game Start" then SPINS FOREVER — see Gotchas #3
```

No CI config, no lint, no `gofmt`/`go vet` enforcement, no `Makefile`.

Test files actually present: `core/actor_test.go` (1 test: `TestNewActor`), `core/storage/entity_map_store_test.go` (5 tests, table-driven, good coverage of the store), `core/timeline_test.go` (1 test: `TestEndToEnd` — the most meaningful test in the repo, walks a full tick schedule), `engine/engine_test.go` (empty `TestEndToEnd` with a TODO comment — **no engine-level test exists**).

## Gotchas & Verified Facts

1. **The engine cannot actually decide actions for actors that need them.** `core.Actor.GetNextAction()` returns an *error* (`ErrActionNotFoundForActor`) when an actor has zero actions — exactly the "needs an action" condition. `engine.getActorsThatNeedActions` appends an actor **only** when `err == nil && actionID == ""`. These two never line up: `err == nil` only happens when `len(actionIDs) >= 1` (in which case `actionID != ""`), and `actionID == ""` only happens when `err` is non-nil (in which case the function returns the error immediately, aborting the whole turn). **Empirically verified**: wiring a store with one `core.NewActor` and a resolver that returns `(true, nil)` causes `RunCurrentTurn()` to return the error `"action not found for actor"` and **no action is ever created**. The `DoNothing` decider is dead code in this engine.

2. **The infinite loop in `example/main.go` is a *consequence* of #1, not a resolver failure.** The example has no actors (only a store + timeline), so `GetAllActors` returns `[]`, `decideActions` produces zero new actions, `GetPendingActionIDWithMaxTick` immediately hits `ErrNoPendingAction`, `ExecuteToTick` returns `nil`, `RunCurrentTurn` advances the tick, and `Run` loops forever. Verified: `timeout 3 go run example/main.go` prints `Game Start` and is killed at 3s with exit 124.

3. **A resolver returning `(false, nil)` is a *fatal error*, not a "retry on next tick."** `engine.resolveAction` does `if !resolved { return fmt.Errorf("action %v was not resolved", actionID) }`. There is no retry, no skip, no fallback. `*action_resolvers.NoAction` always returns `(false, nil)`, so it **cannot** be the last resolver in any chain that actually reaches `resolveAction`. The engine hard-fails the turn.

4. **`Ordered` short-circuits.** `Ordered.ResolveAction` iterates its `ActionResolver` slice; the first one returning `(true, _)` wins and the rest are skipped. Any one returning an error aborts the loop with that error. The first returning `(false, nil)` just moves on.

5. **`SetAttribute` only accepts `string`, `int64`, `float64`, `bool` at runtime** (a `switch` statement in `entity_map_store.go:43-52`). Everything else — `uint64`, `nil`, a struct, a slice — returns the error `Type %v not supported`. This is enforced at the store; there is no compile-time enforcement anywhere in the repo (the `core.Attribute` constraint is declared but never applied).

6. **All values are compared with `==` inside `GetEntitiesWithAttribute`/`GetEntitiesWithAttributes`.** This is only safe because the store restricts values to the four comparable types. Attempting to store a `[]byte` would be rejected by `SetAttribute`, but a `map[string]T` or slice would compile into `SetAttribute` and then **panic** on the `==` comparison if you ever query for it.

7. **The timeline entity ID is *hardcoded* to `"ENTITY_TIMELINE"` for writes only.** `Timeline.SetCurrentTick` unconditionally writes `CURRENT_TICK` onto the entity whose ID is the string `"ENTITY_TIMELINE"`. `Timeline.GetCurrentTick()` reads from *whichever* entity in the store currently has a `CURRENT_TICK` attribute — it does **not** specifically check for the `"ENTITY_TIMELINE"` ID. So: (a) if you manually `store.SetAttribute("myOtherEntity", "CURRENT_TICK", 42)`, `GetCurrentTick` will happily return 42; (b) if **two** entities carry `CURRENT_TICK`, `GetCurrentTick` errors with `"expected 1 or 0 timelines but got 2"`; (c) if no entity carries it, it errors with `"no timeline found, no current tick set"`.

8. **`GetEntitiesWithAttributeType` is O(n) over the entire backing map.** It walks every key, splits on the first `#` via `strings.SplitN(key, "#", 2)`, and collects entity IDs whose attribute portion matches. `RemoveEntity` is also O(n) the same way. `getKey` is `fmt.Sprintf("%s#%s", entityID, attributeID)` — so **entity IDs and attribute IDs must not contain `#`** or the key/parse round-trip is broken.

9. **An "entity" is not a first-class object in the store.** There is no `CreateEntity`, no `ListEntities`, no per-entity record. An entity only *exists* in the sense of appearing as the ID prefix of one or more keys. `RemoveEntity` deletes all keys with that prefix; if you never set any attribute, there is nothing to delete.

10. **`engine.New` is not an error return** — it just fills a struct. Misconfiguration (e.g. a store that doesn't actually satisfy the interface) is a **compile-time** error, not a runtime one. Because the interface is structural, any type that happens to implement the required methods will compile in — including one with subtly wrong behavior — and that surfaces at the first `RunCurrentTurn` call.

11. **`core.ErrNoPendingAction` is the sentinel the engine uses to *exit* `ExecuteToTick` cleanly**, and it is the *only* error path that terminates normally (i.e., without error). Any other error from `GetPendingActionIDWithMaxTick`, `decideActions`, or `resolveAction` propagates up and aborts `Run`.

12. **`action_resolvers.NoAction` is a struct with zero fields and a `ResolveAction` that is a literal `return false, nil`.** It is *not* a no-op resolver — it is a poison resolver. The name is misleading.

13. **`do_nothing` decider's `timeline` field is never used.** `NewDoNothingDecider(entityStore, timeline)` stores it; `DecideActionForActor` only creates the action and returns its ID. The `timeline` parameter is a dead field.

14. **`Action.FinishAction()` = `store.RemoveAttribute(actionID, "ACTION_START_TICK")`** — this is what actually "un-pends" an action (the timeline no longer sees it). `Timeline.RemoveAction(action)` does the same. There is no other way to deschedule.

15. **`Go 1.26` module. Two indirect deps, neither directly imported by any source file:**
    - `github.com/google/uuid v1.3.0` — *is* imported in `core/entity.go:3` for `uuid.New()` (so it's actually a direct dep that is mis-tagged `// indirect`).
    - `github.com/patrickmn/go-cache v2.1.0+incompatible` — **not imported anywhere in the repo.** It is a leftover.

16. **`example/main.go` is runnable but never exits.** It prints `Game Start` and spins (see #2). Do not `Ctrl-C`-expect it to finish. `go test ./...` passes. `go run example/main.go` hangs.

17. **There is no "state" to preserve across turns.** Every turn re-derives everything from the store. The engine can shut down and restart from a snapshot of the store as long as the tick is also in the store (on whichever entity holds `CURRENT_TICK`).
