package minecraftserver

import (
	"context"

	mcspv1 "mcsp.com/server-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Initializing struct{}

func (Initializing) State() mcspv1.ServerState {
	return mcspv1.Initializing
}

func (Initializing) CheckConditions(c *Conditions) bool {
	return !storageExists(c.storage) &&
		!podExists(c.runner) &&
		!podExists(c.uploader)
}

func (i Initializing) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	desiredPVC, err := r.constructStorage(c.ms)
	if err != nil {
		log.Error(err, "failed to construct Storage")
		return err
	}
	err = r.Create(ctx, desiredPVC)
	if err != nil {
		log.Error(err, "unable to create Storage")
		return err
	}
	c.storage = desiredPVC
	log.V(1).Info("created new storage", "PVC", desiredPVC.Name)

	if !controllerutil.ContainsFinalizer(c.ms, minecraftServerFinalizer) {
		controllerutil.AddFinalizer(c.ms, minecraftServerFinalizer)
		if err := r.Update(ctx, c.ms); err != nil {
			log.Error(err, "failed to add finalizer")
			return err
		}
		log.V(1).Info("added finalizer")
		return nil
	}
	return ErrNoAction
}

type Idling struct{}

func (Idling) State() mcspv1.ServerState {
	return mcspv1.Idling
}

func (Idling) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		!podExists(c.runner) &&
		!podExists(c.uploader) &&
		!serverNeedsBackup(c.ms) &&
		!terminating(c)
}

func (i Idling) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	if c.ms.Spec.Active {
		desiredRunner, err := r.constructRunner(c.ms, c.storage)
		if err != nil {
			log.Error(err, "failed to construct Runner")
			return err
		}
		err = r.Create(ctx, desiredRunner)
		if err != nil {
			log.Error(err, "unable to create Runner")
			return err
		}
		log.V(1).Info("created new Runner", "Pod", desiredRunner.Name)
		c.runner = desiredRunner
	}
	return ErrNoAction
}

type Activating struct{}

func (Activating) State() mcspv1.ServerState {
	return mcspv1.Activating
}

func (Activating) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		(podExists(c.runner) && podPending(c.runner)) &&
		!podExists(c.uploader)
}

func (a Activating) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	return ErrNoAction
}

type Running struct{}

func (Running) State() mcspv1.ServerState {
	return mcspv1.Running
}

func (Running) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		(podExists(c.runner) && podRunning(c.runner) && !podTerminating(c.runner)) &&
		!podExists(c.uploader)
}

func (ru Running) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)
	runnerName := c.runner.Name

	annotated := false
	if _, ok := c.ms.Annotations[ServerNeedsBackupAnnotation]; !ok {
		annotated = true
		c.ms.Annotations[ServerNeedsBackupAnnotation] = ServerNeedsBackupAnnotationYes
	}
	if _, ok := c.ms.Annotations[ServerStorageInitAnnotation]; !ok {
		annotated = true
		c.ms.Annotations[ServerStorageInitAnnotation] = ServerStorageInitAnnotationYes
	}
	if annotated {
		if err := r.Update(ctx, c.ms); err != nil {
			log.Error(err, "failed to annotate")
			return err
		}
		log.V(1).Info("annotated")
		return ErrNeedsRequeue
	}

	if !c.ms.Spec.Active || terminating(c) {
		log.V(1).Info("deleting runner", "pod", runnerName)
		if err := r.Delete(ctx, c.runner); err != nil {
			log.Error(err, "failed to delete runner", "pod", runnerName)
			return err
		}
		log.V(1).Info("deleted runner", "pod", runnerName)
	}
	return ErrNoAction
}

type Deactivating struct{}

func (Deactivating) State() mcspv1.ServerState {
	return mcspv1.Deactivating
}

func (Deactivating) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		(podExists(c.runner) && podTerminating(c.runner)) &&
		!podExists(c.uploader)
}

func (Deactivating) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	return ErrNoAction
}

type BackingUp struct{}

func (BackingUp) State() mcspv1.ServerState {
	return mcspv1.BackingUp
}

func (BackingUp) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		!podExists(c.runner) &&
		!podSucceeded(c.uploader) &&
		serverNeedsBackup(c.ms)
}

func (BackingUp) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
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
	}
	return ErrNoAction
}

type CleaningUp struct{}

func (CleaningUp) State() mcspv1.ServerState {
	return mcspv1.CleaningUp
}

func (CleaningUp) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		!podExists(c.runner) &&
		(podExists(c.uploader) && podSucceeded(c.uploader))
}

func (CleaningUp) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	if serverNeedsBackup(c.ms) {
		delete(c.ms.Annotations, ServerNeedsBackupAnnotation)
		if err := r.Update(ctx, c.ms); err != nil {
			log.Error(err, "failed to remove NeedsBackup annotation")
			return err
		}
		log.V(1).Info("removed NeedsBackup annotation")
		return ErrNeedsRequeue
	}
	if controllerutil.ContainsFinalizer(c.uploader, minecraftServerFinalizer) {
		controllerutil.RemoveFinalizer(c.uploader, minecraftServerFinalizer)
		if err := r.Update(ctx, c.uploader); err != nil {
			log.Error(err, "failed to remove Finaler from uploader")
			return err
		}
		return ErrNeedsRequeue
	}
	if !podTerminating(c.uploader) {
		if err := r.Delete(ctx, c.uploader); err != nil {
			log.Error(err, "failed to delete uploader", "pod", c.uploader.Name)
			return err
		}
		log.V(1).Info("deleted uploader", "pod", c.uploader.Name)
		return ErrNeedsRequeue
	}
	return ErrNoAction
}

type Deleting struct{}

func (Deleting) State() mcspv1.ServerState {
	return mcspv1.Deleting
}

func (Deleting) CheckConditions(c *Conditions) bool {
	return storageExists(c.storage) &&
		!podExists(c.runner) &&
		!podExists(c.uploader) &&
		!serverNeedsBackup(c.ms) &&
		terminating(c)
}

func (Deleting) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	log := log.FromContext(ctx)

	if controllerutil.ContainsFinalizer(c.storage, minecraftServerFinalizer) {
		controllerutil.RemoveFinalizer(c.storage, minecraftServerFinalizer)
		if err := r.Update(ctx, c.storage); err != nil {
			log.Error(err, "failed to remove Finalizer from PVC")
			return err
		}
		log.V(1).Info("removed Finalizer from PVC")
		return ErrNeedsRequeue
	}

	if err := r.Delete(ctx, c.storage); err != nil {
		log.Error(err, "failed to delete Storage")
		return err
	}
	log.V(1).Info("deleted Storage")
	if controllerutil.ContainsFinalizer(c.ms, minecraftServerFinalizer) {
		controllerutil.RemoveFinalizer(c.ms, minecraftServerFinalizer)
		if err := r.Update(ctx, c.ms); err != nil {
			log.Error(err, "failed to remove Finalizer")
			return err
		}
		log.V(1).Info("removed Finalizer")
	}
	return nil
}

type Error struct{}

func (Error) State() mcspv1.ServerState {
	return mcspv1.Error
}

func (Error) CheckConditions(c *Conditions) bool {
	return true
}

func (Error) Action(ctx context.Context, r *Reconciler, c *Conditions) error {
	return ErrNoAction
}
