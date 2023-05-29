package minecraftserver

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mcspv1 "mcsp.com/server-operator/api/v1"
)

const (
	PodTypeLabel         = "mcsp.com/pod-role"
	PodTypeLabelRunner   = "runner"
	PodTypeLabelUploader = "uploader"
)

func (r *Reconciler) setupPodOwnerIndexer(mgr ctrl.Manager) error {
	return mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&corev1.Pod{},
		ownerVirtualKey,
		func(rawObj client.Object) []string {
			pod := rawObj.(*corev1.Pod)
			owner := metav1.GetControllerOf(pod)
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

func podExists(pod *corev1.Pod) bool {
	return pod != nil
}
func podRunning(pod *corev1.Pod) bool {
	return pod != nil && pod.Status.Phase == corev1.PodRunning
}
func podTerminating(pod *corev1.Pod) bool {
	return pod != nil && !pod.ObjectMeta.DeletionTimestamp.IsZero()
}

func (r *Reconciler) getRunner(ctx context.Context, ms *mcspv1.MinecraftServer) (*corev1.Pod, error) {
	var pods corev1.PodList
	if err := r.List(ctx, &pods, client.InNamespace(ms.Namespace), client.MatchingFields{
		ownerVirtualKey: ms.Name,
	}, client.MatchingLabels{
		PodTypeLabel: PodTypeLabelRunner,
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

func (r *Reconciler) constructRunner(ms *mcspv1.MinecraftServer, pvc *corev1.PersistentVolumeClaim) (*corev1.Pod, error) {
	storageVolumeName := "storage"
	zipVolumeName := "zip-storage"
	terminationGracePeriodSeconds := int64(120)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-server", ms.Name),
			Namespace: ms.Namespace,
			Labels: map[string]string{
				ServerIDLabel: ms.Spec.Server.ID,
				PodTypeLabel:  PodTypeLabelRunner,
			},
			Annotations: make(map[string]string),
		},
		Spec: corev1.PodSpec{
			RestartPolicy:                 "Never",
			TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,

			InitContainers: []corev1.Container{
				{
					Name:            "loader",
					Image:           "localhost:32000/mcsp/server-loader:v1.0.0",
					ImagePullPolicy: "Always",

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
					}},
				},
			},
			Containers: []corev1.Container{{
				Name:            "runner",
				Image:           "localhost:32000/mcsp/server-runner:v1.0.0",
				ImagePullPolicy: corev1.PullAlways,

				Stdin: true,
				Env: []corev1.EnvVar{{
					Name:  "SERVER_ID",
					Value: ms.Spec.Server.ID,
				}, {
					Name:  "SERVER_JAR",
					Value: ms.Spec.Server.Jarfile,
				}, {
					Name:  "HEAP_MAX",
					Value: ms.Spec.Server.HeapMax,
				}, {
					Name:  "HEAP_INIT",
					Value: ms.Spec.Server.HeapInit,
				}},

				Ports: []corev1.ContainerPort{{
					Name:          "tcp-minecraft",
					ContainerPort: 25565,
				}},
				VolumeMounts: []corev1.VolumeMount{{
					Name:      storageVolumeName,
					MountPath: "/minecraft",
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
