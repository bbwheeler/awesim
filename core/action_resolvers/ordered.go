package action_resolvers

type ActionResolver interface {
	ResolveAction(string) (bool, error)
}

type Ordered struct {
	resolvers []ActionResolver
}

func NewOrdered(resolvers []ActionResolver) *Ordered {
	return &Ordered{
		resolvers: resolvers,
	}
}

func (r *Ordered) ResolveAction(actionID string) (bool, error) {

	for _, resolver := range r.resolvers {
		resolved, err := resolver.ResolveAction(actionID)
		if err != nil {
			return false, err
		}
		if resolved {
			return true, nil
		}
	}
	return false, nil
}
