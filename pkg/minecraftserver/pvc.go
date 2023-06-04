package minecraftserver

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mcspv1 "mcsp.com/server-operator/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PVCInitializedAnnotation        = "mcsp.com/initialized"
	PVCInitializedAnnotationSuccess = "Success"
	PVCInitializedAnnotationFail    = "Error"

	PVCBackupTimeAnnotation = "mcsp.com/backed-up"
)

func storageExists(storage *corev1.PersistentVolumeClaim) bool {
	return storage != nil
}
func storageDeleting(storage *corev1.PersistentVolumeClaim) bool {
	return storage != nil && !storage.ObjectMeta.DeletionTimestamp.IsZero()
}

func (r *Reconciler) setupPVCOwnerIndexer(mgr ctrl.Manager) error {
	return mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&corev1.PersistentVolumeClaim{},
		ownerVirtualKey,
		func(rawObj client.Object) []string {
			pvc := rawObj.(*corev1.PersistentVolumeClaim)
			owner := metav1.GetControllerOf(pvc)
			if owner == nil {
				return nil
			}
			if owner.APIVersion != apiGVStr || owner.Kind != "MinecraftServer" {
				return nil
			}
			return []string{
				owner.Name,
			}
		},
	)
}

func (r *Reconciler) getStorage(ctx context.Context, ms *mcspv1.MinecraftServer) (*corev1.PersistentVolumeClaim, error) {
	var pvcs corev1.PersistentVolumeClaimList
	if err := r.List(ctx, &pvcs, client.InNamespace(ms.Namespace), client.MatchingFields{
		ownerVirtualKey: ms.Name,
	}); err != nil && errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(pvcs.Items) > 0 {
		return &pvcs.Items[0], nil
	}
	return nil, nil
}

func (r *Reconciler) constructStorage(ms *mcspv1.MinecraftServer) (*corev1.PersistentVolumeClaim, error) {
	storageClassName := "kubernetes-work"
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-storage", ms.Name),
			Namespace: ms.Namespace,
			Labels: map[string]string{
				OperatorLabel: OperatorLabelValue,
				ServerIDLabel: ms.Spec.Server.ID,
			},
			Annotations: make(map[string]string),
			Finalizers:  []string{minecraftServerFinalizer},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": ms.Spec.Storage.Capacity,
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(ms, pvc, r.Scheme); err != nil {
		return nil, err
	}
	return pvc, nil
}
