# Extending awesim: Fighter Game Architecture

This document describes how to build a turn-based fighter game on top of awesim in a **separate repository**. It covers what awesim would need to change to support external extensions, and how the external game would be structured.

## Table of Contents

1. [Current State: Why awesim Cannot Be Extended Today](#current-state)
2. [Changes Required in awesim](#changes-required-in-awesim)
3. [External Game Repository Structure](#external-game-repository-structure)
4. [Example: Fighter Game on Top of awesim](#example-fighter-game-on-top-of-awesim)
5. [Module and Import Relationships](#module-and-import-relationships)

---

## <a name="current-state"></a>Current State: Why awesim Cannot Be Extended Today

awesim is **deliberately sealed**. Every type, interface, constant, and function across all packages uses lowercase naming:

| Package | What is Private/Unexported |
|---------|--------------------------|
| `core` | `Entity`, `Actor`, `Action`, `Timeline` (structs), `Engine`, `EntityStore`, `actionDecider`, `actionResolver` (interfaces) |
| `engine` | `RunSimulation`, all internal functions |
| `storage` | `NewMapStore`, `EntityMapStore` |
| `action_deciders` | `DoNothingDecider` |
| Constants | `IsActor`, sentinel attribute keys, `ACTION_INVOKER`, timeline tick key |

No external Go module can:
- Refer to core types (`core.Entity`), the engine interface (`Engine`), or the store interface (`EntityStore`)
- Implement `ActionResolver` or `ActionDecider` (interfaces are unexported)
- Match attribute keys reliably (no exported constants for `IsActor`, `CURRENT_TICK`, etc.)
- Get past the engine loop boundary (no public `Run` method, no tick-by-tick control)

---

## <a name="changes-required-in-awesim"></a>Changes Required in awesim

### 1. Export all core types and interfaces

In `core/entity.go`, `core/actor.go`, `core/action.go`:

```go
// Before (private):
type entity struct { EntityID string }
type actor struct { Entity Entity }
type Action struct { EntityID string; CreatedAt int64 }

// After (public -- uppercase names):
type Entity struct { EntityID string }        // was entity or Actor's embedded field
type Actor struct { Entity Entity; Name string }  // was actor -- make public
type Action struct {                            // already partially public
    ID             string
    CreatedAt      int64
    // add exported fields for external use:
    ActionType     string
    Invokees       []ActorID
}
```

In `core/entity.go` or a new `core/types.go`:

```go
// EntityStore -- currently private interface in core
type EntityStore interface {
    GetAttribute(id string) (interface{}, error)
    SetAttribute(entityID, attributeID string, value interface{}) error
    DeleteAttribute(entityID, attributeID string) error
    DeleteEntity(entityID string) error
}

// Engine -- currently private interface in core
type Engine interface {
    GetEntitiesNeedingActions(store EntityStore) ([]Actor)
    DecideActions(ctx context.Context, store EntityStore, actors []Actor, decider ActionDecider) ([]Action)
    GetNextAction(store EntityStore, actions []Action) (Action, bool)
    ResolveAction(resolvers ActionResolverList, store EntityStore, action Action) error
}

// ActionDecider -- currently private in core/action.go
type ActionDecider interface {
    Decide(ctx context.Context, actor Actor, engine Engine) Action
}

// ActionResolver -- currently private in core/action_resolvers/ordered.go
type ActionResolver interface {
    Resolve(store EntityStore, action Action) (resolved bool, err error)
}
```

### 2. Export all attribute constants

In `core/constants.go` (new file):

```go
package core

// Sentinel attribute keys for discovery and engine behavior.
const (
    IsActorAttribute        = "IS_ACTOR"
    CurrentTickAttribute    = "CURRENT_TICK"
    TimelineEntityID        = "ENTITY_TIMELINE"
    ActionEntityID          = "__ACTION_ENTITY_ID__"
    ActionActionType        = "ACTION_ACTION_TYPE"
    ActionInvokeesAttribute = "ACTION_INVOKEES"
    ActionInvokerAttribute  = "ACTION_INVOKER"
    ActionDurationAttribute = "ACTION_DURATION"
    ActionStartTickAttribute = "ACTION_START_TICK"
)

// Common fighter attribute keys for external games.
// These are not in awesim itself but serve as documented conventions.
const (
    // Shared positioning convention
    PositionXAttribute     = "POSITION_X"
    PositionYAttribute     = "POSITION_Y"

    // Fighter stats (documented, not enforced by awesim)
    HealthAttribute        = "HEALTH"
    AttackPowerAttribute   = "ATTACK_POWER"
    DefenseAttribute       = "DEFENSE"

    // Fighter state
    IsDeadAttribute        = "IS_DEAD"
    CurrentTurnAttribute   = "CURRENT_TURN"
    CanMoveAttribute       = "CAN_MOVE"
    CanActAttribute        = "CAN_ACT"

    // Action types for fighters (game convention, not awesim)
    MoveActionType         = "MOVE"
    PunchActionType        = "PUNCH"
    KickActionType         = "KICK"
)
```

### 3. Export the storage implementation

In `storage/` -- rename everything to start with uppercase:

```go
// Before: package storage; func NewMapStore() EntityMapStore; type entityMapStore struct{...}
// After:
package storage

func NewEntityMapStore() *EntityMapStore

type EntityMapStore struct { /* ...same internals, now public fields or unexported with public methods*/}

func (s *EntityMapStore) GetAttribute(entityID, attributeID string) (interface{}, error) { ... }
func (s *EntityMapStore) SetAttribute(entityID, attributeID string, value interface{}) error { ... }
func (s *EntityMapStore) DeleteAttribute(entityID, attributeID string) error { ... }
func (s *EntityMapStore) DeleteEntity(entityID string) error { ... }
```

### 4. Export the engine entrypoint and tick loop

In `engine/Engine.go`:

```go
// Before: func RunSimulation(engine Engine, store EntityStore) error
// After:

package engine

func Start(ctx context.Context, e Engine, store storage.EntityMapStore, resolvers engine.ActionResolverList) error {
    ticker := time.NewTicker(100 * time.Millisecond) // optional visual update cadence
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            if err := e.Tick(ctx, store); err != nil {
                return err
            }
        }
    }
}

// Tick runs a single simulation step. External games can call Tick() directly per-frame or per-real-time-interval.
func (b *builder) Tick(ctx context.Context, store storage.EntityMapStore) error {
    // existing decision + resolution logic from RunSimulation
}
```

The key change: **expose `Tick()`** so the external game can:
- Call it once per real-time second for a turn-based UI loop
- Inject its own rendering/audio between ticks
- Control when the simulation progresses

### 5. Export helper functions in core/actor.go and core/timeline.go

```go
// In core/actor.go -- currently these are private (lowercase):
func NewActor(store storage.EntityMapStore, actorID string) error
func ActorExists(store storage.EntityMapStore, actorID string) bool // or GetActorByID
func GetAllActors(store storage.EntityMapStore) ([]Actor, error)

// In core/timeline.go -- currently lowercase constants:
const (
    DefaultTimelineEntityID = TimelineEntityID  // re-export as convenient alias
)

func SetCurrentTick(store storage.EntityMapStore, tick int64) error
func GetCurrentTick(store storage.EntityMapStore) (int64, error)
```

### 6. Export the NoActionResolver utility

In `core/action_resolvers/no_action.go`:

```go
// Before: func NoAction(entityStore EntityStore, action Action) (bool, error)
// After:
func NoAction(store storage.EntityMapStore, action Action) (resolved bool, err error) { ... }
```

---

## <a name="external-game-repository-structure"></a>External Game Repository Structure

### Directory layout

```
github.com/myusername/fighter-game/        <-- separate repository, separate Go module
├── go.mod                                  module github.com/myusername/fighter-game
├── go.sum
├── main.go                                CLI entrypoint: game setup + engine Start()
├── fighters/                              Fighter-specific types and behavior
│   ├── fighter.go                          Fighter actor factory, stat methods
│   ├── fighter_decider.go                  ActionDecider for fighters (Move/Punch/Kick logic)
│   ├── fighter_resolvers.go                actionResolver implementations for fighters
│   └── arena.go                            Arena/wiring: creates fighters, sets up game state
├── engine/                                Optional: custom engine extensions
│   └── timer.go                           Timer component (time-limited turns)
├── actions/                              Fighter-specific action definitions
│   ├── move_action.go                      "Move" action definition/conventions
│   ├── punch_action.go                     "Punch" action definition/conventions
│   └── kick_action.go                      "Kick" action definition/conventions
├── gamestate/                            Persistence/snapshot
│   └── snapshot_store.go                   (implements storage.EntityMapStore wrapper or replacement)
└── README.md
```

---

## <a name="example-fighter-game-on-top-of-awesim"></a>Example: Fighter Game on Top of awesim

### `go.mod`

```
module github.com/myusername/fighter-game

go 1.23

require (
    github.com/bbwheeler/awesim v0.0.0
    github.com/google/uuid v1.6.0
)
```

### `fighters/fighter.go` -- Creating fighter actors

```go
package fighters

import (
    "github.com/bbwheeler/awesim/core"
    "github.com/bbwheeler/awesim/storage"
)

// New creates a fighter actor in the store with default stats.
func New(s *storage.EntityMapStore, id string, name string) error {
    if err := core.NewActor(s, id); err != nil {
        return err
    }
    s.SetAttribute(id, core.HealthAttribute, 100)       // fighter health: 100 HP
    s.SetAttribute(id, core.AttackPowerAttribute, 15)   // base damage per attack
    s.SetAttribute(id, core.DefenseAttribute, 5)         // damage reduction
    s.SetAttribute(id, core.PositionXAttribute, 0)       // default grid position
    s.SetAttribute(id, core.PositionYAttribute, 0)
    s.SetAttribute(id, core.CanMoveAttribute, true)      // can move this turn
    s.SetAttribute(id, core.CanActAttribute, true)       // can use a special attack this turn

    _ = name
    return nil
}

// GetProperty reads a fighter attribute. Returns the zero value for the type
// if the attribute doesn't exist (graceful fallback).
func GetProperty(s *storage.EntityMapStore, fighterID, attrKey string) int64 {
    v, err := s.GetAttribute(fighterID, attrKey)
    if err != nil {
        return 0
    }
    switch val := v.(type) {
    case int64:
        return val
    case float64:
        return int64(val)
    default:
        return 0
    }
}

func GetHealth(s *storage.EntityMapStore, fighterID string) int64 {
    return GetProperty(s, fighterID, core.HealthAttribute)
}

func GetPosition(s *storage.EntityMapStore, fighterID string) (int64, int64) {
    x := GetProperty(s, fighterID, core.PositionXAttribute)
    y := GetProperty(s, fighterID, core.PositionYAttribute)
    return x, y
}

func SetPosition(s *storage.EntityMapStore, fighterID string, x, y int64) error {
    if err := s.SetAttribute(fighterID, core.PositionXAttribute, x); err != nil {
        return err
    }
    return s.SetAttribute(fighterID, core.PositionYAttribute, y)
}

func SetProperty(s *storage.EntityMapStore, fighterID, attrKey string, val int64) error {
    return s.SetAttribute(fighterID, attrKey, val)
}
```

### `fighters/fighter_decider.go` -- The fighter action decider

This is the game's primary logic: decides what each fighter does on their turn.

```go
package fighters

import (
    "context"
    "math/rand"
    "github.com/bbwheeler/awesim/core"
)

// FighterDecider implements core.ActionDecider for the fighter game.
// On each tick, it looks at all actors that need actions and gives them fights/attacks/moves.
type FighterDecider struct {
    // Optional: set these to control behavior; if nil they use defaults above.
    DefaultMoveRange int64  // e.g., 3 tile grid distance per turn
    DefaultDuration  int64  // simulation ticks for the action (e.g., 1 tick = 1 second of game time)
}

// Decide is called by awesim's Engine for each actor that needs an action.
func (d *FighterDecider) Decide(ctx context.Context, actor core.Actor, engine core.Engine) core.Action {
    // Skip dead fighters.
    if GetDead(engine.GetStore(), actor.ID) {
        return nil  // no actions for dead entities
    }

    // Simple turn-based decision logic with some randomness:
    switch rand.Intn(3) {
    case 0:
        return CreateMoveAction(ctx, actor)
    case 1:
        return CreatePunchAction(ctx, actor)
    case 2:
        return CreateKickAction(ctx, enemy := FindClosestActor(engine.GetStore(), actor))
    default:
        return nil  // skip this turn (idle)
    }
}

// DecideAll is called internally by the game engine loop. It collects actions from
// all fighters that need them and queues them via awesim's timeline.
func (d *FighterDecider) DecideAll(ctx context.Context, actors []core.Actor, engine core.Engine) ([]core.Action, error) {
    var actions []core.Action
    for _, actor := range actors {
        if action := d.Decide(ctx, actor, engine); action != nil {
            actions = append(actions, action)
        }
    }
    return actions, nil
}

func FindClosestActor(store storage.EntityMapStore, fighter string) core.Actor {
    // Find an enemy within grid range. Returns the nearest alive enemy actor.
    all := GetAllActors(store)  // from awesim/core/actor.go (now exported)
    fx, fy := GetPosition(store, fighter)
    bestDist := int64(1000)   // far
    var closest core.Actor

    for _, candidate := range all {
        if candidate.ID == fighter || core.GetHealth(store, candidate.ID) <= 0 {
            continue
        }
        cx, cy := GetPosition(store, candidate.ID)
        dx := abs(cx - fx)
        dy := abs(cy - fy)
        dist := (dx + dy) // Manhattan distance; could use Euclidean: math.Sqrt(dx*dx+dy*dy)
        if dist < bestDist {
            bestDist = dist
            closest = candidate  // or just store its ID, since Action only needs a reference
        }
    }
    return closest
}

func abs(x int64) int64 {
    if x < 0 { return -x }
    return x
}
```

### `fighters/move_action.go` -- Move action resolver

```go
package fighters

import (
    "github.com/bbwheeler/awesim/core"
    "github.com/bbwheeler/awesim/storage"
)

// CreateMoveAction creates a MOVE-type action entity for the given actor.
func CreateMoveAction(ctx context.Context, actor core.Actor) core.Action {
    // Build an action: awesim's timeline will accept it if SetAttribute is called.
    return core.Action{
        ActionID:     "fighter_move_" + actor.ID,  // convention: type_prefix + "_" + invoker_id
        ActionType:   core.MoveActionType,
        Invoker:      actor.ID,
    }
}

// MoveResolver handles MOVE actions during action resolution.
func MoveResolver(store storage.EntityMapStore, action core.Action) (resolved bool, err error) {
    if action.GetAttribute(core.ActionActionType) != core.MoveActionType {
        return false, nil  // this resolver doesn't handle this action type
    }

    invoker := action.GetInvoker()  // fighter ID
    x, y := GetPosition(store, invoker)

    // Move toward closest enemy (simple deterministic logic).
    target := FindClosestActor(store, invoker)
    if target.ID == "" {
        return true, nil  // no one to move toward; action completes without effect
    }
    tx, ty := GetPosition(store, target.ID)

    dx := sign(tx - x)
    dy := sign(ty - y)
    newX := clamp(x+dx, -20, 20)
    newY := clamp(y+dy, -20, 20)
    SetPosition(store, invoker, newX, newY)

    return true, nil  // move action complete
}

func sign(diff int64) int64 {
    if diff > 0 {
        return 1
    } else if diff < 0 {
        return -1
    }
    return 0
}

func clamp(v, lo, hi int64) int64 {
    if v < lo { return lo }
    if v > hi { return hi }
    return v
}
```

### `fighters/punch_action.go` -- Punch action resolver

```go
package fighters

import (
    "github.com/bbwheeler/awesim/core"
    "github.com/bbwheeler/awesim/storage"
)

func CreatePunchAction(_ context.Context, actor core.Actor) core.Action {
    return core.Action{
        ActionID:   "fighter_punch_" + actor.ID,
        ActionType: core.PunchActionType,
        Invoker:    actor.ID,
    }
}

// PunchResolver resolves punch actions: deals attack_power - opponent defense damage.
func PunchResolver(store storage.EntityMapStore, action core.Action) (resolved bool, err error) {
    if action.GetAttribute(core.ActionActionType) != core.PunchActionType {
        return false, nil
    }

    invoker := action.GetInvoker()  // attacker fighter ID
    target := FindClosestActor(store, invoker)

    if target.ID == "" {
        return true, nil  // no one to punch; action completes
    }

    x1, y1 := GetPosition(store, invoker)
    x2, y2 := GetPosition(store, target.ID)
    dist := (x1-x2)*2 + (y1-y2)*2
    if abs(dist) > 1 {
        return true, nil  // too far to punch; action completes without damage
    }

    attackPower := core.GetAttribute(store, invoker, core.AttackPowerAttribute).(int64)
    defense := core.GetAttribute(store, target.ID, core.DefenseAttribute).(int64)
    damage := max(attackPower-defense, 1)  // minimum 1 damage

    currentHealth := GetHealth(store, target.ID)
    SetProperty(store, target.ID, core.HealthAttribute, currentHealth-damage)

    if currentHealth-damage <= 0 {
        SetProperty(store, target.ID, core.IsDeadAttribute, true)  // mark as dead
    }

    return true, nil
}
```

### `fighters/kick_action.go` -- Kick action resolver

```go
package fighters

import (
    "github.com/bbwheeler/awesim/core"
)

func CreateKickAction(_ context.Context, actor core.Actor) core.Action {
    return core.Action{
        ActionID:   "fighter_kick_" + actor.ID,
        ActionType: core.KickActionType,
        Invoker:    actor.ID,
    }
}

// KickResolver resolves kick actions: higher damage than punch but longer cooldown.
func KickResolver(store storage.EntityMapStore, action core.Action) (resolved bool, err error) {
    if action.GetAttribute(core.ActionActionType) != core.KickActionType {
        return false, nil
    }

    invoker := action.GetInvoker()  // attacker fighter ID
    target := FindClosestActor(store, invoker)

    if target.ID == "" {
        return true, nil
    }

    x1, y1 := GetPosition(store, invoker)
    x2, y2 := GetPosition(store, target.ID)
    dist := (x1-x2)*(abs(x1-x2)+1) + (y1-y2)*(abs(y1-y2)+1)  // Manhattan distance check
    if abs(dist) > 2 {
        return true, nil  // kick range is 2 tiles; too far
    }

    attackPower := core.GetAttribute(store, invoker, core.AttackPowerAttribute).(int64)
    defense := core.GetAttribute(store, target.ID, core.DefenseAttribute).(int64)
    damage := max(attackPower*2-defense, 1)  // kick does 2x attack power before defense

    currentHealth := GetHealth(store, target.ID)
    SetProperty(store, target.ID, core.HealthAttribute, currentHealth-damage)

    if currentHealth-damage <= 0 {
        SetProperty(store, target.ID, core.IsDeadAttribute, true)
    }

    return true, nil
}
```

### `main.go` -- Game entrypoint: wiring everything together

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/bbwheeler/awesim/core"
    "github.com/bbwheeler/awesim/engine"
    "github.com/bbwheeler/awesim/storage"
    "github.com/myusername/fighter-game/fighters"
)

func main() {
    // Create the backing store.
    store := storage.NewEntityMapStore()

    // Seed timeline (required: must exist before engine runs).
    SetCurrentTick(store, 1)

    // Create two fighters.
    fighters.New(store, "fighter_1", "Iron Fist")
    fighters.New(store, "fighter_2", "Shadow Fist")

    // Build the FighterDecider.
    decider := &fighters.FighterDecider{DefaultMoveRange: 3, DefaultDuration: 1}

    // Build the resolver chain -- awesim's first matching resolver wins.
    var resolvers []core.ActionResolver {
        fighters.NoActionResolver(store),     // debugging: instant-resolve all (uncomment to test setup)
        fighters.PunchResolver,               // resolve PUNCH actions
        fighters.KickResolver,                // resolve KICK actions
        fighters.MoveResolver,                // resolve MOVE actions
    }

    // Build the engine with the decider and resolvers.
    e := engine.NewEngine(decider, resolvers)

    // Run the simulation loop (cancellable by Ctrl+C).
    ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer cancel()

    log.Println("Starting fighter simulation...")
    if err := engine.Start(ctx, e, store); err != nil {
        log.Fatalf("Simulation ended: %v", err)
    }
    log.Println("Simulation complete.")
}
```

---

## <a name="module-and-import-relationships"></a>Module and Import Relationships

```
fighter-game/                          fighter-game/
├── go.mod                             ├── fighters/
│   require(                           │   ├── fighter.go            (Actor creation, stat getters/setters)
│     awesim/core                      │   ├── fighter_decider.go    (ActionDecider implementation)
│     awesim/engine                    │   ├── fighter_resolvers.go  (MoveResolver, PunchResolver, KickResolver)
│   )                                  │   └── arena.go              (Arena setup/wiring)
│                                      ├── actions/
├── main.go                           │   ├── move_action.go        CreateMoveAction()
│                                     │   ├── punch_action.go       CreatePunchAction()
│                                     │   └── kick_action.go        CreateKickAction()
fighters go.mod                          ├── fighters/               (see above)
require(                               ├── engine/timer.go           (turn timer component)
  awesim/core                            └── gamestate/snapshot_store.go
    awesim/engine                      └── main.go                  (entrypoint)
  awesim/storage                        )
)                                      fighters/
                                        go.mod
                                        require(
                                            github.com/bbwheeler/awesim/*
```

**Two separate repos:**

| Repository | Module name | Purpose | Depends on awesim? |
|------------|-------------|---------|-------------------|
| `github.com/bbwheeler/awesim` | `github.com/bbwheeler/awesim` | ECS engine + core types | No (dependency-free) |
| `github.com/myusername/fighter-game` | `github.com/myusername/fighter-game` | Fighter game on top of awesim | Yes: `awesim/core`, `awesim/engine`, `awesim/storage` |

---

## Summary of What Would Need to Change in awesim

| # | Change | Effort |
|---|--------|--------|
| 1 | Export every type, interface, and function that was previously lowercase (uppercase names) | Medium |
| 2 | Add `core/constants.go` with all attribute sentinel keys as exported const values | Small |
| 3 | Rename `entityMapStore` -> `EntityMapStore`, export constructor to `NewEntityMapStore` | Small |
| 4 | Export `Engine.Tick()` for tick-by-tick control from external games | Small |
| 5 | Move `RunSimulation` to a public `Start(ctx, Engine, store)` that wraps Tick() in a loop | Medium |
| 6 | Ensure all parameter types in exported interfaces use exported names (not internal) | Medium |
| 7 | Document the attribute key conventions so external games know what keys are safe to use | Small |

## Key Design Principle

The core insight: **awesim provides the ECS runtime; the game provides the semantics.** awesim should have zero knowledge of "fighters", "health", "moves", or "attacks". Those lives entirely in the external repo. awesim's job is to:

1. Store entity attributes (key-value pairs)
2. Discover actors via sentinel attributes
3. Schedule actions on a timeline
4. Ask deciders what each actor should do
5. Run resolvers to apply action effects
6. Advance ticks between resolved actions

Everything else -- fighter creation, combat math, grid positions, health bars, UI rendering -- is the external game's responsibility.
