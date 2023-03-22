package core

const actionInvoker string = "ACTION_INVOKER"
const actionDuration string = "ACTION_DURATION"

type Action struct {
	Entity
}

func AsAction(e *Entity) *Action {
	return &Action{
		Entity: *e,
	}
}


func NewAction(invoker *Actor, duration int64, dao EntityDao) (*Action, error) {
	entity := NewEntity(invoker.dao)
	entity.SetAttribute(actionInvoker, invoker.GetID())
	entity.SetAttribute(actionDuration, duration)
	return &Action{
		Entity: *entity,
	}, nil
}

func (a *Action) GetInvoker() (*Actor, error) {
	invokerID, err := a.GetAttribute(actionInvoker)
	if err != nil {
		return nil, err
	}
	actorEntity := GetEntity(invokerID.(string), a.dao)
	return &Actor{
		Entity: *actorEntity,
	}, nil
}
func (a *Action) GetDuration() (int64, error) {
	duration, err := a.GetAttribute(actionDuration)
	return duration.(int64), err
}

func (a *Action) FinishAction() error {
	return a.dao.RemoveAttribute(a.GetID(), actionStartTick)
}