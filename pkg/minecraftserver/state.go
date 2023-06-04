package minecraftserver

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	mcspv1 "mcsp.com/server-operator/api/v1"
)

type Conditions struct {
	ms       *mcspv1.MinecraftServer
	storage  *corev1.PersistentVolumeClaim
	runner   *corev1.Pod
	uploader *corev1.Pod
}

type State interface {
	State() mcspv1.ServerState
	CheckConditions(c *Conditions) bool
	Action(context.Context, *Reconciler, *Conditions) error
}

var ErrNeedsRequeue = errors.New("needs requeue")
var ErrNoAction = errors.New("no action")

func IsNoAction(err error) bool {
	return err.Error() == ErrNoAction.Error()
}
func IsNeedsRequeue(err error) bool {
	return err.Error() == ErrNeedsRequeue.Error()
}
