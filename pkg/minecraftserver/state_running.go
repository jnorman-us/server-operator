package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Running struct{}

func (Running) State() mcspv1.ServerState {
	return mcspv1.Running
}

func (Running) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		(podExists(c.runner) && podRunning(c.runner) && !podTerminating(c.runner)) &&
		!podExists(c.uploader)
}

func (Running) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)
	runnerName := c.runner.Name

	if _, ok := c.ms.Annotations[ServerNeedsBackupAnnotation]; !ok {
		c.ms.Annotations[ServerNeedsBackupAnnotation] = ServerNeedsBackupAnnotationYes
		if err := r.Update(ctx, c.ms); err != nil {
			log.Error(err, "failed to annotate as NeedsBackup")
			return err
		}
		log.V(1).Info("annotated as NeedsBackup")
	}
	if !c.ms.Spec.Active || serverTerminating(c.ms) {
		log.V(1).Info("deleting runner", "pod", runnerName)
		if err := r.Delete(ctx, c.runner); err != nil {
			log.Error(err, "failed to delete runner", "pod", runnerName)
			return err
		}
		log.V(1).Info("deleted runner", "pod", runnerName)
		return nil
	}
	return ErrNoAction
}
