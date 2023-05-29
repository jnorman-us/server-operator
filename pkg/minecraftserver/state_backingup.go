package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type BackingUp struct{}

func (b BackingUp) State() mcspv1.ServerState {
	return mcspv1.BackingUp
}

func (b BackingUp) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		!podExists(c.runner) &&
		!uploaderCompleted(c.uploader) &&
		serverNeedsBackup(c.ms)
}

func (b BackingUp) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	if !podExists(c.uploader) {
		desiredUploader, err := r.constructUploader(c.ms, c.storage)
		if err != nil {
			log.Error(err, "failed to construct Uploader")
			return err
		}
		err = r.Create(ctx, desiredUploader)
		if err != nil {
			log.Error(err, "unable to create Uploader")
			return err
		}
		log.V(1).Info("created new Uploader", "Pod", desiredUploader.Name)
		c.uploader = desiredUploader
		return nil
	}
	return ErrNoAction
}
