package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
)

type Error struct{}

func (e Error) State() mcspv1.ServerState {
	return mcspv1.Error
}

func (e Error) CheckConditions(c *Conditions) bool {
	return true
}

func (e Error) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	return ErrNoAction
}
