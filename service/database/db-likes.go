package database

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/codes"
)

func (db *AppDatabase) LikeExists(ctx context.Context, ownerId int64, photoId int64, userId int64) (bool, error) {

	ctx, span := tracer.Start(ctx, "database.LikeExists")
	defer span.End()

	var count int
	err := db.c.QueryRowContext(ctx, `SELECT COUNT(*) FROM likes WHERE photo_owner = ? AND photo_id = ? AND user_id = ?;`, ownerId, photoId, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return false, fmt.Errorf("can't check if the like exists: %w", err)
	}

	return count > 0, nil
}

func (db *AppDatabase) CreateLike(ctx context.Context, ownerId int64, photoId int64, userId int64) error {

	ctx, span := tracer.Start(ctx, "database.CreateLike")
	defer span.End()

	_, err := db.c.ExecContext(ctx, `INSERT INTO likes (photo_owner, photo_id, user_id) VALUES (?, ?, ?);`, ownerId, photoId, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Insert failed")
		return fmt.Errorf("can't insert the like: %w", err)
	}

	return nil
}

func (db *AppDatabase) DeleteLike(ctx context.Context, ownerId int64, photoId int64, userId int64) error {

	ctx, span := tracer.Start(ctx, "database.DeleteLike")
	defer span.End()

	_, err := db.c.ExecContext(ctx, `DELETE FROM likes WHERE photo_owner = ? AND photo_id = ? AND user_id = ?;`, ownerId, photoId, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete like")
		return fmt.Errorf("can't delete the like: %w", err)
	}

	return nil
}
