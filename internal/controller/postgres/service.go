package postgres

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	databasesv1alpha1 "github.com/nikitadada/atlasdb/api/v1alpha1"
)

func BuildHeadlessService(pg *databasesv1alpha1.PostgresCluster) *corev1.Service {
	labels := Labels(pg.Name)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pg.Name,
			Namespace: pg.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None", // headless
			Selector:  labels,
			Ports: []corev1.ServicePort{
				{
					Name:     "postgres",
					Port:     5432,
					Protocol: corev1.ProtocolTCP,
				},
			},
		},
	}
}
