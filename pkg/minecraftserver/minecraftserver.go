package minecraftserver

import (
	"context"

	ref "k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcspv1 "mcsp.com/server-operator/api/v1"
)

const (
	ServerIDLabel = "mcsp.com/serverId"

	ServerNeedsBackupAnnotation    = "mcsp.com/needs-backup"
	ServerNeedsBackupAnnotationYes = "yes"
	ServerNeedsBackupAnnotationNo  = "no"
)

func serverNeedsBackup(ms *mcspv1.MinecraftServer) bool {
	if value, ok := ms.Annotations[ServerNeedsBackupAnnotation]; ok {
		return value == ServerNeedsBackupAnnotationYes
	}
	return false
}

func (r *Reconciler) updateServerStatus(
	ctx context.Context,
	ms *mcspv1.MinecraftServer,
	state mcspv1.ServerState,
	c *Conditions,
) error {
	log := log.FromContext(ctx)

	if c.storage != nil {
		pvcRef, err := ref.GetReference(r.Scheme, c.storage)
		if err != nil {
			log.Error(err, "unable to get PVC reference")
		}
		ms.Status.Storage.PVCRef = pvcRef
	}
	if c.runner != nil {
		podRef, err := ref.GetReference(r.Scheme, c.runner)
		if err != nil {
			log.Error(err, "unable to get Pod reference")
		}
		ms.Status.Runner.PodRef = podRef
		ms.Status.Runner.PodIP = c.runner.Status.PodIP
	}
	if c.uploader != nil {
		podRef, err := ref.GetReference(r.Scheme, c.uploader)
		if err != nil {
			log.Error(err, "unable to get Pod reference")
		}
		ms.Status.Uploader.PodRef = podRef
	}
	ms.Status.State = state

	return r.Status().Update(ctx, ms)
}
