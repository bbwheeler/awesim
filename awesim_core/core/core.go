package core

const IsActorAttribute string = "IsActor"

type Action interface {
	Entity
	GetInvoker() (Actor, error)
	GetStartTick() (int64, error)
	GetEndTick() (int64, error)
}

type ActionType interface {
	CreateActionForActor(Actor) (Action, error)
}


type Entity interface {
	GetID() string
	GetStringAttribute(attribute string) (string, error) 
	GetNumberAttribute(attribute string) (float64, error) 
	GetIntegerAttribute(attribute string) (int64, error) 
	GetBooleanAttribute(attribute string) (bool, error) 
}

type Actor interface {
	Entity
	GetNextAction() (Action, error)
	GetPossibleActionTypes() ([]ActionType, error)
}

type ActionDecider interface {
	GetActionForActor(actor Actor) (Action, error)
}

type ActionResolver interface {
	ResolveAction(action Action) error
}