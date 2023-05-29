package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
)

type Deactivating struct{}

func (d Deactivating) State() mcspv1.ServerState {
	return mcspv1.Deactivating
}

func (d Deactivating) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		(podExists(c.runner) && podTerminating(c.runner)) &&
		!podExists(c.uploader)
}

func (d Deactivating) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	return ErrNoAction
}
