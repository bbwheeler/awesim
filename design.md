# Design

## Process

### Startup

1. Insert all entities
2. Run Loop

1. Register actions with action resolvers
1. Register Entities

### Loop
 
EntityManager -> GetEntitiesNeedingActions
ActionDecider -> DecideActions
Timeline -> AddActions
Timeline -> GetNextAction
Resolver -> ResolveAction
 -> list of resolvers (in order, first one to resolve resolves)
  -> Resolve

## Items

Entity
* Identifier

Action (sub entity)
* invoker
* start time
* end time
