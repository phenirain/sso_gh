package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/phenirain/sso/pkg/contextkeys"
)

func MustInitDb(cs string) *sqlx.DB {
	db, err := sqlx.Connect("postgres", cs)
	if err != nil {
		panic(err.Error())
	}
	err = db.Ping()
	if err != nil {
		panic(err.Error())
	}
	return db
}

func WithUserTransaction[T any](db *sqlx.DB, ctx context.Context, f func(tx *sqlx.Tx) (T, error)) (T, error) {
	var zero T

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return zero, err
	}

	userID, ok := ctx.Value(contextkeys.UserIDCtxKey).(int)
	if !ok {
		userID = 0
	}

	_, err = tx.ExecContext(ctx, fmt.Sprintf("SET LOCAL myapp.current_user_id = %d", userID))
	if err != nil {
		_ = tx.Rollback()
		return zero, err
	}

	result, err := f(tx)
	if err != nil {
		_ = tx.Rollback()
		return zero, err
	}

	if err := tx.Commit(); err != nil {
		return zero, err
	}

	return result, nil
}
