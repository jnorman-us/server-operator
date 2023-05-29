package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Idling struct{}

func (i Idling) State() mcspv1.ServerState {
	return mcspv1.Idling
}

func (i Idling) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		!podExists(c.runner) &&
		!podExists(c.uploader) &&
		!serverNeedsBackup(c.ms)
}

func (i Idling) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	if c.ms.Spec.Active {
		desiredRunner, err := r.constructRunner(c.ms, c.storage)
		if err != nil {
			log.Error(err, "failed to construct Runner")
			return err
		}
		err = r.Create(ctx, desiredRunner)
		if err != nil {
			log.Error(err, "unable to create Runner")
			return err
		}
		log.V(1).Info("created new Runner", "Pod", desiredRunner.Name)
		c.runner = desiredRunner
		return nil
	}
	return ErrNoAction
}
