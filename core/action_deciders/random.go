package action_deciders

import "github.com/bbwheeler/awesim/core"

type Random struct {
	entityStore core.EntityStore
	timeline    *core.Timeline
}

func (r *Random) DecideActionForActor(actorID string) (string, error) {
	// Step One
	// Get Potential Actions for Actor

	// Step Two
	// Decide on a random action

	// Step Three
	// Fill in the action parameters

	// Step Four
	// Return the action ID

	return "", core.ErrActionNotFoundForActor
}
