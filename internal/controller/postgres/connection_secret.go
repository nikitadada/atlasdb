package postgres

import (
	"context"
	"fmt"
	dbv1alpha1 "github.com/nikitadada/atlasdb/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildConnectionSecret(
	clusterName string,
	namespace string,
	host string,
	password string,
) *corev1.Secret {
	port := "5432"
	username := "postgres"
	database := "postgres"

	uri := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		username,
		password,
		host,
		port,
		database,
	)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName + "-conn",
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"host":     host,
			"port":     port,
			"username": username,
			"password": password,
			"database": database,
			"uri":      uri,
		},
	}
}

func GetPostgresPassword(
	ctx context.Context,
	k8sClient client.Client,
	namespace string,
	secretName string,
) (string, error) {
	secret := &corev1.Secret{}

	err := k8sClient.Get(
		ctx,
		client.ObjectKey{
			Name:      secretName,
			Namespace: namespace,
		},
		secret,
	)
	if err != nil {
		return "", err
	}

	passwordBytes, ok := secret.Data["password"]
	if !ok {
		return "", fmt.Errorf("password key not found in secret")
	}

	return string(passwordBytes), nil
}

func ReconcileConnectionSecret(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	pg *dbv1alpha1.PostgresCluster,
	creds *corev1.Secret,
) error {

	name := pg.Name + "-connection"

	password := string(creds.Data["password"])
	username := "postgres"
	db := "postgres"
	host := pg.Name + "-rw"
	port := "5432"

	uri := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		username, password, host, port, db,
	)

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: pg.Namespace,
		},
		StringData: map[string]string{
			"host":     host,
			"port":     port,
			"username": username,
			"password": password,
			"database": db,
			"uri":      uri,
		},
	}

	err := ctrl.SetControllerReference(pg, &secret, scheme)
	if err != nil {
		return err
	}

	return upsertSecret(ctx, c, &secret)
}
