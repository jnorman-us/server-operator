package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type CleaningUp struct{}

func (cu CleaningUp) State() mcspv1.ServerState {
	return mcspv1.CleaningUp
}

func (cu CleaningUp) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		!podExists(c.runner) &&
		(podExists(c.uploader) && uploaderCompleted(c.uploader))
}

func (cu CleaningUp) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	if serverNeedsBackup(c.ms) {
		delete(c.ms.Annotations, ServerNeedsBackupAnnotation)
		if err := r.Update(ctx, c.ms); err != nil {
			log.Error(err, "failed to remove NeedsBackup annotation")
			return err
		}
		log.V(1).Info("removed NeedsBackup annotation")
	}
	if !podTerminating(c.uploader) {
		if err := r.Delete(ctx, c.uploader); err != nil {
			log.Error(err, "failed to delete uploader", "pod", c.uploader.Name)
			return err
		}
		log.V(1).Info("deleted uploader", "pod", c.uploader.Name)
	}
	return ErrNoAction
}
