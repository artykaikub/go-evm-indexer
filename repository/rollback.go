package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type Rollback struct {
	client *mongo.Client
}

func NewRollback(client *mongo.Client) Rollback {
	return Rollback{
		client: client,
	}
}

// ExecTransaction creates a new transaction and handles rollback/commit based on the
// error object returned by the `function input`
func (r *Rollback) ExecTransaction(ctx context.Context, fn func(sc mongo.SessionContext) error) error {
	err := r.client.UseSession(ctx, func(sc mongo.SessionContext) error {
		if err := sc.StartTransaction(); err != nil {
			return err
		}

		if fn != nil {
			if err := fn(sc); err != nil {
				sc.AbortTransaction(sc)
				return err
			}
		}

		return sc.CommitTransaction(sc)
	})

	return err
}
