package postgres

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func upsertSecret(
	ctx context.Context,
	c client.Client,
	desired *corev1.Secret,
) error {

	var existing corev1.Secret

	err := c.Get(ctx, client.ObjectKey{
		Name:      desired.Name,
		Namespace: desired.Namespace,
	}, &existing)

	// --- create ---
	if apierrors.IsNotFound(err) {
		return c.Create(ctx, desired)
	}

	// --- real error ---
	if err != nil {
		return err
	}

	// --- update ---
	existing.Data = desired.Data
	existing.StringData = desired.StringData
	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations

	return c.Update(ctx, &existing)
}
