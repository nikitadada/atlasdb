package postgres

import (
	"context"

	dbv1alpha1 "github.com/nikitadada/atlasdb/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const PostgresFinalizer = "databases.atlasdb.io/finalizer"

func EnsureFinalizer(
	ctx context.Context,
	c client.Client,
	pg *dbv1alpha1.PostgresCluster,
) error {
	if pg.DeletionTimestamp.IsZero() {
		// объект НЕ удаляется
		if !controllerutil.ContainsFinalizer(pg, PostgresFinalizer) {
			controllerutil.AddFinalizer(pg, PostgresFinalizer)
			return c.Update(ctx, pg)
		}
		return nil
	}

	// объект УДАЛЯЕТСЯ
	if controllerutil.ContainsFinalizer(pg, PostgresFinalizer) {
		// TODO: здесь позже будет backup / cleanup
		controllerutil.RemoveFinalizer(pg, PostgresFinalizer)
		return c.Update(ctx, pg)
	}

	return nil
}
