package controller

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	databasesv1alpha1 "github.com/nikitadada/atlasdb/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
