package minecraftserver

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	mcspv1 "mcsp.com/server-operator/api/v1"
)

var (
	apiGVStr        = mcspv1.GroupVersion.String()
	ownerVirtualKey = ".metadata.ownerMinecraftServer"
)

var states = []State{
	Initializing{},
	Idling{},
	Activating{},
	Running{},
	Deactivating{},
	BackingUp{},
	CleaningUp{},
	Error{},
}

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.setupPodOwnerIndexer(mgr)
	r.setupPVCOwnerIndexer(mgr)
	return ctrl.NewControllerManagedBy(mgr).
		For(&mcspv1.MinecraftServer{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Complete(r)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MinecraftServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	var ms mcspv1.MinecraftServer
	if err := r.Get(ctx, req.NamespacedName, &ms); err != nil {
		log.Error(err, "unable to fetch")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var conditions = Conditions{
		ms: &ms,
	}
	var fetchErr error
	conditions.storage, fetchErr = r.getStorage(ctx, &ms)
	if fetchErr != nil {
		log.Error(fetchErr, "unable to fetch storage PVC")
		return ctrl.Result{}, fetchErr
	}
	conditions.runner, fetchErr = r.getRunner(ctx, &ms)
	if fetchErr != nil {
		log.Error(fetchErr, "unable to fetch runner Pod")
		return ctrl.Result{}, fetchErr
	}
	conditions.uploader, fetchErr = r.getUploader(ctx, &ms)
	if fetchErr != nil {
		log.Error(fetchErr, "unable to fetch uploader Pod")
		return ctrl.Result{}, fetchErr
	}

	var currentState State
	for _, state := range states {
		if state.CheckConditions(&conditions) {
			currentState = state
			break
		}
	}

	actionErr := currentState.Action(ctx, r, &conditions)
	if actionErr != nil && IsNoAction(actionErr) {
		log.V(1).Info("performed no Action", "currentState", currentState.State())
	} else if actionErr != nil {
		log.Error(actionErr, "error performing Action", "currentState", currentState.State())
		return ctrl.Result{}, actionErr
	} else {
		log.V(1).Info("performed Action", "currentState", currentState.State())
	}

	updateStatusError := r.updateServerStatus(ctx, &ms, currentState.State(), &conditions)
	if updateStatusError != nil {
		log.Error(updateStatusError, "failed to update MinecraftServer status")
	}

	return ctrl.Result{}, nil
}
