package core

const IsActorAttribute string = "IsActor"

type ActionPossibilitizer interface {
	GetPotentialActions(actor *Actor) ()
}

type Attribute interface { string|int64|float64|bool}