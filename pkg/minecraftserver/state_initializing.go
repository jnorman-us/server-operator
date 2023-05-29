package minecraftserver

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	mcspv1 "mcsp.com/server-operator/api/v1"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Initializing struct{}

func (i Initializing) State() mcspv1.ServerState {
	return mcspv1.Initializing
}

func (i Initializing) CheckConditions(c *Conditions) bool {
	return !storageExists(c.storage) &&
		!podExists(c.runner) &&
		!podExists(c.uploader)
}

func (i Initializing) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	if !controllerutil.ContainsFinalizer(c.ms, minecraftServerFinalizer) {
		controllerutil.AddFinalizer(c.ms, minecraftServerFinalizer)
		if err := r.Update(ctx, c.ms); err != nil {
			log.Error(err, "failed to add Finalizer")
			return err
		}
		log.V(1).Info("added Finalizer")
	}

	if !storageExists(c.storage) {
		desiredPVC, err := r.constructStorage(c.ms)
		if err != nil {
			log.Error(err, "failed to construct Storage")
			return err
		}
		err = r.Create(ctx, desiredPVC)
		if err != nil {
			log.Error(err, "unable to create Storage")
			return err
		}
		log.V(1).Info("created new storage", "PVC", desiredPVC.Name)
		c.storage = desiredPVC
		return nil
	}
	return ErrNoAction
}
