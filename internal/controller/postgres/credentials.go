package postgres

import (
	"context"

	databasesv1alpha1 "github.com/nikitadada/atlasdb/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ReconcileCredentials(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	pg *databasesv1alpha1.PostgresCluster,
) (*corev1.Secret, error) {

	name := pg.Spec.SuperuserSecretName
	if name == "" {
		name = pg.Name + "-superuser"
	}

	var secret corev1.Secret
	err := c.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: pg.Namespace,
	}, &secret)

	if err == nil {
		return &secret, nil
	}

	if !apierrors.IsNotFound(err) {
		return nil, err
	}

	password := GeneratePassword(32)

	secret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: pg.Namespace,
		},
		StringData: map[string]string{
			"username": "postgres",
			"password": password,
		},
	}

	if err := controllerutil.SetControllerReference(pg, &secret, scheme); err != nil {
		return nil, err
	}

	if err := c.Create(ctx, &secret); err != nil {
		return nil, err
	}

	return &secret, nil
}
