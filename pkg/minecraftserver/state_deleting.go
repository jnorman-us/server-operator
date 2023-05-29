package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Deleting struct{}

func (Deleting) State() mcspv1.ServerState {
	return mcspv1.Deleting
}

func (Deleting) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		!podExists(c.runner) &&
		!podExists(c.uploader) &&
		!serverNeedsBackup(c.ms) &&
		serverTerminating(c.ms)
}

func (Deleting) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	if err := r.Delete(ctx, c.storage); err != nil {
		log.Error(err, "failed to delete Storage")
		return err
	}
	log.V(1).Info("deleted Storage")
	if controllerutil.ContainsFinalizer(c.ms, minecraftServerFinalizer) {
		controllerutil.RemoveFinalizer(c.ms, minecraftServerFinalizer)
		if err := r.Update(ctx, c.ms); err != nil {
			log.Error(err, "failed to remove Finalizer")
			return err
		}
		log.V(1).Info("removed Finalizer")
	}
	return nil
}
