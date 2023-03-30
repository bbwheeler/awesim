package core

const IsActorAttribute string = "IsActor"

type ActionDecider interface {
	DecideActionForActor(actor *Actor) (*Action, error)
}

type ActionResolver interface {
	ResolveAction(action *Action) (bool, error)
}

type ActionPossibilitizer interface {
	GetPotentialActions(actor *Actor) ()
}

type Attribute interface { string|int64|float64|bool}