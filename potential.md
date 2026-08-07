# Potential Improvements for awesim

## System Overview

awesim is a Go-based ECS (Entity-Component-System) game simulation engine that models turn-based simulation via an Entity-Action-Timeline loop. Entities are identified by UUID strings and store keyed attributes in a `core.IsActorAttribute` sentinel pattern. Actions are themselves entity pairs with invoker, duration, and start tick metadata, scheduled on a global timeline entity (`ENTITY_TIMELINE`).

High-level flow per tick:
1. **Find** actors whose next action has finished or don't yet have one
2. **Decide** what action each actor takes (pluggable deciders)
3. **Schedule** actions on the timeline with start ticks
4. **Resolve** the pending action from the timeline that finishes earliest (pluggable resolvers, first-wins-ordered-chain)

---

## Critical Bugs & Stability Issues

### 1. `Engine.Run()` is an infinite loop — no exit condition
`engine/Engine.go:86-94` — The `Run()` method spins forever because `ended` is never set to `true`. Combined with the `NoAction` resolver (which returns `(false, nil)`) and a tick-1 start, this will call `RunCurrentTurn()` infinitely. This is a live fire hazard for consumers who invoke `Run()`.

### 2. `getActorsThatNeedActions` has dead code: checks `actionID == ""`
`engine/Engine.go:137` — After calling `GetNextAction()`, the code checks `err == nil && actionID == ""`, but `GetNextAction()` at `core/actor.go:32-44` returns an error (never empty string) when no action exists. The `actionID == ""` branch is unreachable. The logic only works because of Go's `err != nil` fallthrough, making it confusing to read and maintain.

### 3. `Timeline.GetCurrentTick()` does a full scan every call
`core/timeline.go:44-63` — To get the current tick, it scans all attributes in the map via `GetEntitiesWithAttributeType()`, filtering for `currentTickAttribute`, then checks entity count == 1. This is O(n) over the entire attribute map. For a value that never changes identity, this should be cached or computed directly from `ENTITY_TIMELINE#CURRENT_TICK`.

### 4. Action existence check relies on attributes (ghost actions)
A dead entity with just an `ACTION_START_TICK` attribute can be discovered by `getActions()` via `GetEntitiesWithAttributeType(actionStartTickAttribute)`. If the action itself entity is partially deleted, the timeline can reference a non-action. There's no validation that discovered "action" IDs are actually actions (no ACTION_DURATION check).

### 5. EntityMapStore.GetEntitiesWithAttributes uses O(n^2) intersection for multiple attributes
`core/storage/entity_map_store.go:76-93` — When called with >1 attribute, it iterates the full map once counting matches per entity, then filters by `count == len(attributes)`. This works correctly but is fragile. The `intersection()` helper at line 95 (and its helpers `contains`, `indexOf`, `removeAtIndex`) at lines 115-130 are dead code — they're defined but never called anywhere in the codebase.

### 6. Timeline.GetEntitiesWithAttributeType returns duplicates
`core/timeline.go:45` calls `GetEntitiesWithAttributeType(currentTickAttribute)`, which scans the flat map and extracts entity IDs from keys matching the attribute suffix. Since only one entity should have this attribute, it's fine in practice. But if two entities both set `CURRENT_TICK`, the function returns 2 — causing a double-count timeline bug with no enforcement beyond "expected 1 or 0".

### 7. Action resolver error handling ambiguity
`engine/Engine.go:166-175` — When `(resolved, err) := ResolveAction(...)` returns `err != nil`, the engine returns an error (fails hard). When it returns `(false, nil)`, the engine calls `RemoveEntity`. But `(true, nil)` — resolved successfully — also causes `RemoveEntity`. The contract is clear in code but unclear at the interface level: does a resolver returning `(true, nil)` mean "I cleaned up and you should remove the entity" or "I handled it, stop looking"? The engine always removes the entity on any `resolved=true` path regardless.

### 8. `getActions` has no dedup check against resolved actions
After resolution, the engine calls `entityProvider.RemoveEntity(actionID)`, which deletes all attributes of the action entity. But between `GetPendingActionIDWithMaxTick()` and `RemoveEntity()`, another tick's `decideActions()` could schedule overlapping actions. No concurrency protection exists, so race conditions are possible in multi-threaded consumers.

---

## Efficiency Improvements

### 9. Replace flat map queries with index structures
The entire attribute store is a single `map[string]any` keyed by `"entityID#attributeID"`. Every discovery (`GetEntitiesWithAttributeType`, `GetAllActors`) iterates the full map. Adding an inverted index (`map[attributeName]map[entityID]bool` or similar) would reduce O(n) scans to O(1).

### 10. Attribute key parsing creates allocations
`getEntityIDFromKey()` and `getAttributeFromKey()` at `core/storage/entity_map_store.go:132-140` each call `strings.SplitN(key, "#", 2)` on every query. This is a string split on the hot path. Using `strings.Index()` + slicing or a custom fast-path for the `#` separator avoids unnecessary allocations.

### 11. RemoveEntity does a full map scan
`core/storage/entity_map_store.go:19-26` — To delete an entity, it iterates every key in the attribute map and checks prefix match. This is O(n) where n = total attributes across all entities. A per-entity attribute list or index would make this O(a) where a = attributes of that one entity.

### 12. GetEntitiesWithAttribute creates unnecessary intermediate maps
`core/storage/entity_map_store.go:69-73` — `GetEntitiesWithAttribute(attr, val)` delegates to `GetEntitiesWithAttributes(map[string]any{attr: val})`, which iterates the entire map. For single-attribute queries, direct lookup via an index would be much faster at scale.

---

## Architecture & Design Issues

### 13. No component isolation — everything is a flat attribute bag
The ECS pattern promises component-based composition, but every entity in this system is just `(entityID#attributeID) → value` with no type safety between components (e.g., Position vs Velocity vs Health all share the same attribute map). There's no schema validation — any string can be set as a "component".

### 14. Action entity lifecycle is unclear
Actions are created via `NewAction()`, scheduled via `timeline.AddAction()`, then resolved and removed via `ResolveAction()` + `RemoveEntity()`. But:
- No guarantee that `ACTION_INVOKER` exists (created by NewAction but not validated)
- `FinishAction()` at `core/action.go:42-44` only removes the "tick" attribute, leaving the action entity alive as a ghost
- Actions and entities share the same ID space with no structural distinction

### 15. Engine interfaces duplicate EntityStore — unnecessary indirection
`engine/Engine.go:34-44` — The `entityProvider` interface replicates every method from `EntityStore`. This adds zero abstraction — the engine only needs `GetEntitiesWithAttribute`, not the full store API. A subset interface would be cleaner, or simply inject the `storage.EntityMapStore` directly.

### 16. No dependency injection for timeline storage
`core/timeline.go:15-19` — The Timeline wraps a single `EntityStore` that it uses to read/write its own persistent state (`CURRENT_TICK`). This tightly couples Timeline to the global store and makes unit testing require seeding an entity with `ENTITY_TIMELINE#CURRENT_TICK`. A simpler `Timeline` struct could hold its tick in a field: `type Timeline struct { currentTick int64 }`.

### 17. Action type discriminator doesn't exist
There's no attribute that differentiates "this is an Action entity" from "this is a regular Entity". The engine discovers actions by looking for the presence of `ACTION_START_TICK`, which leaks implementation details and creates false positives if a non-action entity happens to have that attribute name.

### 18. Decider interface only supports one action per actor
`core/action_deciders/do_nothing.go:19-24` — `DecideActionForActor(actorID) (string, error)` returns a single string. If the decider wants to schedule multiple actions for an actor, it must either create and schedule them itself or return no action at all.

### 19. Resolver chain semantics are opaque
The `Ordered` resolver pattern is simple (first to resolve wins), but:
- No way to know which resolver "won" for debugging/instrumentation
- No error propagation convention: when resolver 2 errors, does that mean "this action can't be resolved by this resolver" or "the whole system failed"? The `ResolveAction` at line 22 returns the first error immediately.

### 20. No tick boundary enforcement in engine loop
`engine/Engine.go:97-127` — `ExecuteToTick()` decides actions, finds earliest pending action with max_tick constraint, resolves it, and loops back to decide again. This means every iteration re-scans all actors for action-needing (step 3 of the tick) repeatedly within a single tick's processing, which is O(actors × resolved_actions_per_tick). Most engines batch the action-decision per tick once, then process actions without re-scheduling mid-tick.

---

## Refactoring Opportunities

### 21. `errString` should be exported with an unexported sentinel ( idiomatic Go)
`core/entity.go:6` — `errString` is a local string alias used for sentinel errors (`ErrActionNotFoundForActor`, `ErrNoPendingAction`). In Go, the convention is to use `var ErrXXX = errors.New("...")` at package level with a blank-identifier type check.

### 22. Action/Entity/Actor constructors expose raw storage
All factory functions (`NewEntity`, `NewActor`, `NewAction`, `GetEntity`, etc.) directly reference the store, which means every consumer must import and pass the same store instance — coupling domain concepts to infrastructure. A builder or manager pattern would encapsulate this.

### 23. Type assertions without verification are unsafe
Multiple locations (e.g., `core/action.go:34` with `invokerID.(string)`, `core/action.go:39` with `duration.(int64)`) perform blind type assertions that can panic at runtime if a wrong type was stored. Go's assert-idiom (`val, ok := anyVal.(T)`) is never used.

### 24. `asActor()` and `asAction()` wrappers serve no purpose
`core/actor.go:47-50` and `core/action.go:47-50` — These are identity conversions that add a type layer without adding behavior. They create code duplication (same methods on both `Entity` and embedded subtypes) with no added value beyond namespace organization.

### 25. Naming inconsistencies across the codebase
- `NoAction.go` in action_resolvers has wrong case (`NoAction` vs Go conventions which favor `noaction` for filenames)
- `daoKey`, `daoValue` variable names in storage refer to "DAO pattern" but this is an in-memory attribute store
- `entityMapStore` receiver name uses snake_case while Go convention is camelCase
- `action_deciders` package (snake case dir) vs `action_resolvers` (no underscore in dir but capitalized types)

### 26. `Action.FinishAction()` partially cleans up
`core/action.go:42-44` — Only removes `ACTION_START_TICK`, leaving ACTION_INVOKER, ACTION_DURATION, and other action-related attributes on the entity. This is a half-delete that creates a zombie action entity detectable by `GetEntitiesWithAttributeType(actionStartTickAttribute) == false`.

### 27. `getActions()` untyped slice of strings — no compile safety
`core/timeline.go:146-148` returns `[]string`, which callers must interpret as action IDs or entity IDs interchangeably. A typed wrapper (`type ActionID string`) would prevent mixing at compile time.

### 28. Engine doesn't expose per-tick metrics/status
There's no observability into what happened during a tick: how many actors were scheduled, how many actions resolved, how many failed. In a game simulation engine, this information is essential for debugging and profiling.

---

## Missing Features

### 29. No test coverage on the engine loop
`engine/engine_test.go:5-6` — Only contains `// TODO: End to end test`. The core business logic (decide actions → schedule → resolve in a multi-turn loop) is completely untested. This is the highest-risk part of the codebase with zero tests.

### 30. No entity lifecycle management
There's no way to "destroy" an entity (it was on `RemoveEntity()` but no public API exposes it for destruction after creation). Once an actor exists, nothing can de-activate or clean it up.

### 31. Actions have no priority/ordering beyond end tick
All pending actions are ordered by end tick only. If multiple actions finish at the same tick, behavior is non-deterministic (depends on map iteration order). Games need priority resolution for simultaneous events.

### 32. No event system or side-effects
Resolved actions produce no observable side effects beyond entity deletion. There's no hook for "something happened" beyond the actor receiving a new action. A proper ECS should have systems that observe and react to state changes (renderers, AI planners, physics calculators all react to data).

### 33. `go.mod` includes unused dependency
`go.mod:7` — `github.com/patrickmn/go-cache v2.1.0+incompatible` is listed as a dependency but never imported anywhere in the codebase. It adds ~KB of bloat with no value.

### 34. No system/world abstraction
The engine directly manipulates entities. A proper ECS separates the "world" (which holds all component data and manages entity creation/destruction) from "systems" (which process entities matching a query). This separation would enable parallel systems, filtering, and multiple worlds.

---

## Code Quality Notes

### 35. Error wrapping loses context
Errors flow up through many layers without adding context: `core/actor.go:38` ("expected 1 action for invoker but got N") → `engine/Engine.go:169` (returned as-is). Adding `fmt.Errorf("...: %w", err)` at each layer would preserve stack context for debugging.

### 36. The timeline is the only hardcoded entity
`core/timeline.go:5` — The timeline uses a string constant `"ENTITY_TIMELINE"` as its ID, making it discoverable in the same global attribute store but giving it no special treatment (no wrapper that caches the ID, for example).

### 37. Engine.RunCurrentTurn doesn't use `context.Context`
No cancellation or timeout is supported. Any caller could hang forever with an infinite loop if all actors keep scheduling zero-duration actions. A context parameter would be the standard Go pattern for this.

### 38. No benchmark suite
There's no performance testing to validate whether the ECS works at scale (10k entities, concurrent systems, etc.). The O(n) full-scan store patterns identified above can only be validated with benchmarks.

---

## Priority Summary

| Category | High Priority | Medium Priority | Low Priority |
|----------|--------------|-----------------|--------------|
| **Stability** | #1 (infinite loop), #7 (resolve semantics), #8 (concurrency) | #4 (ghost actions), #33 (unused dep) | #6, #32, #37 |
| **Efficiency** | #9 (indexes), #10 (key parsing), #12 (intermediate maps) | #11 (RemoveEntity), #13-#14 (component isolation) | #28 (metrics) |
| **Refactoring** | #23 (type assertions), #25 (naming), #125 (interfaces) | #16 (timeline storage), #20 (tick design), #21 (error types) | #17, #24, #27 |
| **Testing** | #29 (engine test coverage) | #30-#31 (entity lifecycle, priorities) | #38 (benchmarks) |