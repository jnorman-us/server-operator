package minecraftserver

import (
	"context"

	ref "k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcspv1 "mcsp.com/server-operator/api/v1"
)

const (
	OperatorLabel      = "mcsp.com/operator"
	OperatorLabelValue = "server-operator"

	ServerIDLabel = "mcsp.com/serverId"

	ServerNeedsBackupAnnotation    = "mcsp.com/needs-backup"
	ServerNeedsBackupAnnotationYes = "yes"
	ServerNeedsBackupAnnotationNo  = "no"

	ServerStorageInitAnnotation    = "mcsp.com/storage-init"
	ServerStorageInitAnnotationYes = "yes"
	ServerStorageInitAnnotationNo  = "no"
)

func serverStorageInitialized(ms *mcspv1.MinecraftServer) bool {
	if value, ok := ms.Annotations[ServerStorageInitAnnotation]; ok {
		return value == ServerStorageInitAnnotationYes
	}
	return false
}
func serverNeedsBackup(ms *mcspv1.MinecraftServer) bool {
	if value, ok := ms.Annotations[ServerNeedsBackupAnnotation]; ok {
		return value == ServerNeedsBackupAnnotationYes
	}
	return false
}
func terminating(c *Conditions) bool {
	return c.ms != nil && !c.ms.ObjectMeta.DeletionTimestamp.IsZero() ||
		storageDeleting(c.storage)
}

func (r *Reconciler) constructStatus(
	ctx context.Context,
	state mcspv1.ServerState,
	c *Conditions,
) error {
	log := log.FromContext(ctx)
	status := mcspv1.MinecraftServerStatus{
		State: state,
	}
	if storageExists(c.storage) {
		pvcRef, err := ref.GetReference(r.Scheme, c.storage)
		if err != nil {
			log.Error(err, "unable to get PVC reference")
		}
		status.Storage.PVCRef = pvcRef
	}
	if podExists(c.runner) {
		runnerRef, err := ref.GetReference(r.Scheme, c.runner)
		if err != nil {
			log.Error(err, "unable to get Pod reference")
		}
		status.Runner.PodRef = runnerRef
		status.Runner.PodIP = c.runner.Status.PodIP
	}
	if podExists(c.uploader) {
		uploaderRef, err := ref.GetReference(r.Scheme, c.uploader)
		if err != nil {
			log.Error(err, "unable to get Pod reference")
		}
		status.Uploader.PodRef = uploaderRef
	}
	c.ms.Status = status
	if err := r.Status().Update(ctx, c.ms); err != nil {
		log.Error(err, "failed to update status")
		return err
	}
	log.V(1).Info("updated status")
	return nil
}
