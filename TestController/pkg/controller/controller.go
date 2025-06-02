package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PodReconciler struct {
	client.Client
}

// Reconcile gets called when a Pod is created, updated, or deleted.
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		// If the Pod no longer exists, no requeue is needed.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling Pod", "name", pod.Name, "namespace", pod.Namespace)

	// Your logic here — for example, annotate the Pod if not already annotated
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	if _, ok := pod.Annotations["reconciled"]; !ok {
		pod.Annotations["reconciled"] = "true"
		if err := r.Update(ctx, &pod); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("Annotated pod with reconciled=true")
	}

	return ctrl.Result{}, nil
}

func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}
