package database

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/codes"
)

var ErrFollowYourself = errors.New("can't follow yourself")

func (db *AppDatabase) FollowExists(ctx context.Context, userId int64, followUserId int64) (bool, error) {

	ctx, span := tracer.Start(ctx, "database.FollowExists")
	defer span.End()

	var count int
	err := db.c.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE user_id = ? AND followed_user = ?;`, userId, followUserId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return false, fmt.Errorf("query error: %w", err)
	}

	return count > 0, nil
}

func (db *AppDatabase) CreateFollow(ctx context.Context, userId int64, followUserId int64) error {

	ctx, span := tracer.Start(ctx, "database.CreateFollow")
	defer span.End()

	// Check if the user is trying to follow itself
	if userId == followUserId {
		return ErrFollowYourself
	}

	// Insert the follow
	_, err := db.c.ExecContext(ctx, `INSERT INTO follows (user_id, followed_user) VALUES (?, ?);`, userId, followUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Insert failed")
		return fmt.Errorf("db insert error: %w", err)
	}

	return nil
}

func (db *AppDatabase) DeleteFollow(ctx context.Context, userId int64, followUserId int64) error {

	ctx, span := tracer.Start(ctx, "database.DeleteFollow")
	defer span.End()

	// Delete the follow
	_, err := db.c.ExecContext(ctx, `DELETE FROM follows WHERE user_id = ? AND followed_user = ?;`, userId, followUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete failed")
		return fmt.Errorf("db delete error: %w", err)
	}

	return nil
}
