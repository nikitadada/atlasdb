package postgres

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	dbv1alpha1 "github.com/nikitadada/atlasdb/api/v1alpha1"
)

func BuildStatefulSet(
	cluster *dbv1alpha1.PostgresCluster,
) *appsv1.StatefulSet {

	labels := map[string]string{
		"app": cluster.Name,
	}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: pointer.Int32(cluster.Spec.Instances),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			ServiceName: cluster.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "postgres",
							Image: "postgres:15",
							Ports: []corev1.ContainerPort{
								{ContainerPort: 5432},
							},
						},
					},
				},
			},
		},
	}
}
