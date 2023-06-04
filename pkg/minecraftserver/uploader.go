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

func (r *Reconciler) getUploader(ctx context.Context, ms *mcspv1.MinecraftServer) (*corev1.Pod, error) {
	var pods corev1.PodList
	if err := r.List(ctx, &pods, client.InNamespace(ms.Namespace), client.MatchingFields{
		ownerVirtualKey: ms.Name,
	}, client.MatchingLabels{
		PodTypeLabel: PodTypeLabelUploader,
	}); err != nil && errors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	if len(pods.Items) > 0 {
		return &pods.Items[0], nil
	}
	return nil, nil
}

func (r *Reconciler) constructUploader(ms *mcspv1.MinecraftServer, pvc *corev1.PersistentVolumeClaim) (*corev1.Pod, error) {
	storageVolumeName := "storage"
	zipVolumeName := "zip-storage"
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-uploader", ms.Name),
			Namespace: ms.Namespace,
			Labels: map[string]string{
				OperatorLabel: OperatorLabelValue,
				ServerIDLabel: ms.Spec.Server.ID,
				PodTypeLabel:  PodTypeLabelUploader,
			},
			Annotations: make(map[string]string),
			Finalizers:  []string{minecraftServerFinalizer},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Containers: []corev1.Container{{
				Name:            "uploader",
				Image:           "localhost:32000/mcsp/uploader:v1.0.2",
				ImagePullPolicy: corev1.PullAlways,
				Env: []corev1.EnvVar{{
					Name:  "SERVER_ID",
					Value: ms.Spec.Server.ID,
				}, {
					Name:  "STORAGE_URL",
					Value: "simple-upload-server.uploads.svc.cluster.local:25478",
				}, {
					Name:  "STORAGE_TOKEN",
					Value: "c68rutn6bhubhiqwertjun",
				}},
				VolumeMounts: []corev1.VolumeMount{{
					Name:      storageVolumeName,
					MountPath: "/minecraft",
				}, {
					Name:      zipVolumeName,
					MountPath: "/zip",
				}},
			}},
			Volumes: []corev1.Volume{{
				Name: storageVolumeName,
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc.Name,
					},
				},
			}, {
				Name: zipVolumeName,
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			}},
		},
	}
	pod.ObjectMeta.Labels[ServerIDLabel] = ms.Spec.Server.ID
	if err := ctrl.SetControllerReference(ms, pod, r.Scheme); err != nil {
		return nil, err
	}
	return pod, nil
}
