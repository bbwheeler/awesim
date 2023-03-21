package action_resolvers

import "github.com/bbwheeler/awesim/awesim_core/core"

type Ordered struct {
	resolvers []core.ActionResolver
}

func NewOrdered(resolvers []core.ActionResolver) *Ordered {
	return &Ordered{
		resolvers: resolvers,
	} 
}

func (r *Ordered)ResolveAction(action *core.Action) (bool, error) {

	for _, resolver := range r.resolvers {
		resolved, err := resolver.ResolveAction(action)
		if err != nil {
			return false, err
		}
		if resolved {
			return true, nil
		}
	}
	return false, nil 
}
