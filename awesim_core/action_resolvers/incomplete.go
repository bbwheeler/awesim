package action_resolvers

import "github.com/bbwheeler/awesim/awesim_core/core"

type Incomplete struct {

}

func (ar *Incomplete)ResolveAction(action core.Action) error {
	return nil // TODO: No action resolver is complete
}
