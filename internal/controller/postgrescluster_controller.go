package controller

import (
	"context"
	databasesv1alpha1 "github.com/nikitadada/atlasdb/api/v1alpha1"
	"github.com/nikitadada/atlasdb/internal/controller/postgres"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PostgresClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=databases.atlasdb.io,resources=postgresclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.atlasdb.io,resources=postgresclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.atlasdb.io,resources=postgresclusters/finalizers,verbs=update

func (r *PostgresClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling PostgresCluster", "name", req.NamespacedName)

	pg := &databasesv1alpha1.PostgresCluster{}
	if err := r.Get(ctx, req.NamespacedName, pg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := postgres.EnsureFinalizer(ctx, r.Client, pg); err != nil {
		return ctrl.Result{}, err
	}

	if !pg.ObjectMeta.DeletionTimestamp.IsZero() {
		// объект удаляется, дальше reconcile не нужен
		return ctrl.Result{}, nil
	}

	var sts appsv1.StatefulSet
	err := r.Get(ctx, types.NamespacedName{
		Name:      pg.Name,
		Namespace: pg.Namespace,
	}, &sts)

	if apierrors.IsNotFound(err) {
		desired := postgres.BuildStatefulSet(pg)

		if err := ctrl.SetControllerReference(pg, desired, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}
	}

	pg.Status.Phase = "Reconciling"
	if err := r.Status().Update(ctx, pg); err != nil {
		return ctrl.Result{}, err
	}

	meta.SetStatusCondition(&pg.Status.Conditions, metav1.Condition{
		Type:    "Ready",
		Status:  metav1.ConditionFalse,
		Reason:  "Reconciling",
		Message: "Postgres cluster is being reconciled",
	})

	if err := r.Status().Update(ctx, pg); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PostgresClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&databasesv1alpha1.PostgresCluster{}).
		Named("postgrescluster").
		Complete(r)
}
