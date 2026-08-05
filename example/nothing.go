package main

import "github.com/bbwheeler/awesim/core"

type ActionStore interface {
	NewAction(invokerID string, duration core.Tick) (*core.Entity, error)
}

type Nothing struct {
	actionStore ActionStore
}

func NewNothing(actionStore ActionStore) *Nothing {
	return &Nothing{
		actionStore: actionStore,
	}
}

func (r *Nothing) ResolveAction(actionID string) error {
	// Do nothing
	return nil
}

func (p *Nothing) ProvideNextActionFor(actorID string) (string, error) {
	action, err := p.actionStore.NewAction(actorID, 1)
	if err != nil {
		return "", err
	}

	return action.GetID(), nil
}
