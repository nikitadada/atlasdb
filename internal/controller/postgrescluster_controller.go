package controller

import (
	"context"
	"fmt"
	"time"

	databasesv1alpha1 "github.com/nikitadada/atlasdb/api/v1alpha1"
	"github.com/nikitadada/atlasdb/internal/controller/postgres"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PostgresClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const FinalizerName = "databases.atlasdb.io/finalizer"

// +kubebuilder:rbac:groups=databases.atlasdb.io,resources=postgresclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=databases.atlasdb.io,resources=postgresclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=databases.atlasdb.io,resources=postgresclusters/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *PostgresClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling PostgresCluster", "name", req.NamespacedName)

	pg := &databasesv1alpha1.PostgresCluster{}
	if err := r.Get(ctx, req.NamespacedName, pg); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// ---------------- FINALIZER ----------------

	if err := postgres.EnsureFinalizer(ctx, r.Client, pg); err != nil {
		return ctrl.Result{}, err
	}

	creds, err := postgres.ReconcileCredentials(ctx, r.Client, r.Scheme, pg)
	if err != nil {
		return ctrl.Result{}, err
	}

	// ---------------- STATEFULSET ENSURE ----------------

	sts := &appsv1.StatefulSet{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      pg.Name,
		Namespace: pg.Namespace,
	}, sts)

	if apierrors.IsNotFound(err) {
		desired := postgres.BuildStatefulSet(pg)

		if err := ctrl.SetControllerReference(pg, desired, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Creating StatefulSet")
		if err := r.Create(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}

		// ждём следующий reconcile
		return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// ---------------- SERVICE ENSURE ----------------

	svc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      pg.Name,
		Namespace: pg.Namespace,
	}, svc)

	if apierrors.IsNotFound(err) {
		desiredSvc := postgres.BuildHeadlessService(pg)

		if err := ctrl.SetControllerReference(pg, desiredSvc, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Creating Headless Service")

		if err := r.Create(ctx, desiredSvc); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	err = postgres.ReconcileConnectionSecret(ctx, r.Client, r.Scheme, pg, creds)
	if err != nil {
		return ctrl.Result{}, err
	}

	// ---------------- CLIENT SERVICE ENSURE ----------------

	clientSvcName := pg.Name + "-rw"

	clientSvc := &corev1.Service{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      clientSvcName,
		Namespace: pg.Namespace,
	}, clientSvc)

	if apierrors.IsNotFound(err) {
		desired := postgres.BuildClientService(pg)

		if err := ctrl.SetControllerReference(pg, desired, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		logger.Info("Creating client service")

		if err := r.Create(ctx, desired); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// =========================
	// CREATE CONNECTION SECRET
	// =========================

	password, err := postgres.GetPostgresPassword(
		ctx,
		r.Client,
		pg.Namespace,
		pg.Spec.SuperuserSecretName,
	)
	if err != nil {
		return ctrl.Result{}, err
	}

	host := pg.Name + "-rw"

	connSecret := BuildConnectionSecret(
		pg.Name,
		pg.Namespace,
		host,
		password,
	)

	// set owner reference
	if err := controllerutil.SetControllerReference(
		pg,
		connSecret,
		r.Scheme,
	); err != nil {
		return ctrl.Result{}, err
	}

	existing := &corev1.Secret{}
	err = r.Get(
		ctx,
		client.ObjectKey{
			Name:      connSecret.Name,
			Namespace: connSecret.Namespace,
		},
		existing,
	)

	if err != nil && apierrors.IsNotFound(err) {
		err = r.Create(ctx, connSecret)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if pg.Status.ConnectionSecret != connSecret.Name {
		pg.Status.ConnectionSecret = connSecret.Name
		err = r.Status().Update(ctx, pg)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// ---------------- READINESS CHECK ----------------

	if sts.Status.ReadyReplicas != *sts.Spec.Replicas {
		logger.Info("Waiting for StatefulSet to be ready",
			"ready", sts.Status.ReadyReplicas,
			"desired", *sts.Spec.Replicas)

		meta.SetStatusCondition(&pg.Status.Conditions, metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionFalse,
			Reason:             "Reconciling",
			Message:            "Waiting for pods to become ready",
			LastTransitionTime: metav1.Now(),
		})

		pg.Status.Phase = "Reconciling"

		if err := r.Status().Update(ctx, pg); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	// ---------------- READY ----------------

	logger.Info("Postgres cluster is ready")

	meta.SetStatusCondition(&pg.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "ClusterReady",
		Message:            "Postgres cluster is ready",
		LastTransitionTime: metav1.Now(),
	})

	pg.Status.Phase = "Ready"
	pg.Status.Endpoint = fmt.Sprintf(
		"%s.%s.svc.cluster.local:5432",
		clientSvcName,
		pg.Namespace,
	)

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
