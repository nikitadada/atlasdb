package controller

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
