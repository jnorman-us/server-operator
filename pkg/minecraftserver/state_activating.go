package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
)

type Activating struct{}

func (a Activating) State() mcspv1.ServerState {
	return mcspv1.Activating
}

func (a Activating) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		(podExists(c.runner) && !podRunning(c.runner) && !podTerminating(c.runner)) &&
		!podExists(c.uploader)
}

func (a Activating) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	return ErrNoAction
}
