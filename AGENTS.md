# AGENTS.md

## Architecture

**awesim** is a Go ECS (Entity-Component-System) game simulation engine. It models turn-based simulation via an Entity-Action-Timeline loop.

### Package boundaries

| Package | Purpose |
|---------|---------|
| `core` | Core types: `Entity`, `Actor`, `Action`, `Timeline`. All public. |
| `core/storage` | In-memory attribute storage (`EntityMapStore`). Single backing map keyed by `entityID#attributeID`. |
| `core/action_deciders` | Pluggable action deciders (currently only `DoNothingDecider`). |
| `core/action_resolvers` | Pluggable action resolvers (`NoAction`, `Ordered` composite). |
| `engine` | Simulation loop: decides which actors need actions, schedules them on the timeline, resolves pending actions. |
| `example` | Minimal wiring for the simulation engine. Not a library — just demo setup. |

### Entity model

- Entities are identified by UUID strings (via `github.com/google/uuid`).
- All data is attribute-keyed: an entity is just a set of `(attributeKey, value)` pairs stored in an `EntityStore`.
- Actorness is a sentinel attribute: `core.IsActorAttribute` (value `IsActor`, bool `true`). Use `core.GetAllActors(store)` to list them.
- Actions are entities with `ACTION_INVOKER` and `ACTION_DURATION` attributes. When scheduled, the timeline also sets `ACTION_START_TICK`.
- The timeline itself is an entity with ID `"ENTITY_TIMELINE"` storing a `CURRENT_TICK` attribute.

### Engine loop (engine/Engine.go)

1. `GetEntitiesNeedingActions` — finds all actors whose next action has finished or doesn't exist yet.
2. `DecideActions` — calls the action decider for each actor, adds resulting actions to the timeline.
3. `GetNextAction` — picks the pending action with earliest end tick.
4. `ResolveAction` — first resolver in chain wins; removes the action entity from storage.

## Commands

```bash
go mod tidy            # add/update dependencies
go test ./...          # run all tests
go run example/main.go # run the demo (sets tick=1, spins Run() via NoAction resolver)
```

No lint, typecheck, fmt, or CI config is present. Tests are standard `testing` package with only unit specs on storage and actor creation — no engine-level integration tests exist yet (`engine_test.go` has a pending TODO).

## Gotchas

- **Attribute types**: `EntityMapStore.SetAttribute` only accepts `string`, `int64`, `float64`, `bool`. Passing anything else (e.g. `uint64`, `nil`) returns an error at runtime.
- **Key format**: Attributes are stored as `entityID#attributeID` strings in one flat map. An entity with no attributes doesn't exist in the store — discovery relies on querying attribute values rather than listing entities.
- **Timeline must be seeded**: The engine panics if no timeline entity exists. Call `timeline.SetCurrentTick(1)` before running.
- **Action resolving is atomic**: When a resolver returns `(false, nil)`, the action remains pending and will be retried on the next tick. Returning `(true, err)` or `(false, err)` are both handled — check your resolver's error semantics carefully.
- **Ordered resolver short-circuits**: The first resolver that returns `resolved=true` wins; subsequent resolvers in the chain are never called for that action.

## Beginning Tasks
Before you begin, follow these steps:
1. Make sure all existing changes have been checked in; if there are existing changes, commit them and push them to git.wheeli.ca.
2. Do a git fetch so that you have all of the latest changes.
3. Switch to a branch or create a branch appropriate for the changes that you will make

## Finishing Tasks
Once you complete any changes, additions, deletions, or modifications, follow these steps:
1. Check the code into a branch using git
2. Push the code to git.wheeli.ca
3. Open a Pull Request for the changes you just pushed
4. Add me (brian) as a reviewer on the Pull Request

## Credentials
Your credentials for git.wheeli.ca can be found in the parent directory (../credentials.md)