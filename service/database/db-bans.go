package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/codes"
)

var ErrBanYourself = errors.New("can't ban yourself")

func (db *AppDatabase) BanExists(ctx context.Context, userId int64, bannedUserId int64) (bool, error) {

	ctx, span := tracer.Start(ctx, "database.BanExists")
	defer span.End()

	var count int
	err := db.c.QueryRowContext(ctx, `SELECT COUNT(*) FROM bans WHERE user_id = ? AND banned_user = ?;`, userId, bannedUserId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return false, fmt.Errorf("query error: %w", err)
	}

	return count > 0, nil
}

// CreateBan registers a ban in the database. If the banned user was following the user, the follow is deleted.
// It returns an error if the user is trying to ban itself.
func (db *AppDatabase) CreateBan(ctx context.Context, userId int64, bannedUserId int64) error {

	ctx, span := tracer.Start(ctx, "database.CreateBan")
	defer span.End()

	// Check if the user is trying to ban itself
	if userId == bannedUserId {
		return ErrBanYourself
	}

	// Create a transaction
	tx, err := db.c.Begin()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Starting transaction failed")
		return fmt.Errorf("can't start a transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// Check if the banned user was following the user
	var count int
	err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM follows WHERE user_id = ? AND followed_user = ?;`, bannedUserId, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return fmt.Errorf("can't check if the banned user was following the user: %w", err)
	}

	// If so, delete the follow
	if count > 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM follows WHERE user_id = ? AND followed_user = ?;`, bannedUserId, userId)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Delete failed")
			return fmt.Errorf("can't delete the follow: %w", err)
		}
	}

	// Insert the ban
	_, err = tx.ExecContext(ctx, `INSERT INTO bans (user_id, banned_user) VALUES (?, ?);`, userId, bannedUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Insert failed")
		return fmt.Errorf("db insert error: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Commit failed")
		return fmt.Errorf("can't commit the transaction: %w", err)
	}

	return nil
}

func (db *AppDatabase) DeleteBan(ctx context.Context, userId int64, bannedUserId int64) error {

	ctx, span := tracer.Start(ctx, "database.DeleteBan")
	defer span.End()

	// Delete the follow
	_, err := db.c.ExecContext(ctx, `DELETE FROM bans WHERE user_id = ? AND banned_user = ?;`, userId, bannedUserId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete failed")
		return fmt.Errorf("db delete error: %w", err)
	}

	return nil
}
