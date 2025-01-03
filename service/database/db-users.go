package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel/codes"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameAlreadyExists = errors.New("username already exists")

// GetUserId queries the database for the user ID of the given username. If the user is not found, it returns an errorUserNotFound.
func (db *AppDatabase) GetUserId(ctx context.Context, username string) (int64, error) {

	ctx, span := tracer.Start(ctx, "database.GetUserId")
	defer span.End()

	var id int64

	err := db.c.QueryRowContext(ctx, `SELECT user_id FROM users WHERE username = ?;`, username).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, ErrUserNotFound
	} else if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return -1, fmt.Errorf("can't get the user ID: %w", err)
	}

	return id, nil
}

// CreateUser creates a new user with the given username. It returns the user ID of the new user.
func (db *AppDatabase) CreateUser(ctx context.Context, username string) (int64, error) {

	ctx, span := tracer.Start(ctx, "database.CreateUser")
	defer span.End()

	// Insert the new user into the database
	res, err := db.c.ExecContext(ctx, `INSERT INTO users(username) VALUES (?);`, username)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Insert failed")
		return -1, fmt.Errorf("can't create the user: %w", err)
	}

	// Get the user ID of the new user
	id, err := res.LastInsertId()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get the user ID")
		return -1, fmt.Errorf("can't get the user ID: %w", err)
	}

	return id, nil
}

// UserExists checks if the user ID exists in the database.
func (db *AppDatabase) UserExists(ctx context.Context, userId int64) (bool, error) {

	ctx, span := tracer.Start(ctx, "database.UserExists")
	defer span.End()

	var id int64

	err := db.c.QueryRowContext(ctx, `SELECT user_id FROM users WHERE user_id = ?;`, userId).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return false, fmt.Errorf("can't check if the user ID exists: %w", err)
	}

	return true, nil
}

// SetUsername sets the username of the user with the given user ID.
// It returns an errorUserNotFound if the user is not found, and an errorUsernameAlreadyExists if the new username already exists.
func (db *AppDatabase) SetUsername(ctx context.Context, userID int64, newUsername string) error {

	ctx, span := tracer.Start(ctx, "database.SetUsername")
	defer span.End()

	// Check if the new username already exists
	_, err := db.GetUserId(ctx, newUsername)
	if !errors.Is(err, ErrUserNotFound) {
		if err == nil {
			return ErrUsernameAlreadyExists
		} else {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to check if the username exists")
			return fmt.Errorf("can't check if the username already exists: %w", err)
		}
	}

	// Update the username of the user
	res, err := db.c.ExecContext(ctx, `UPDATE users SET username = ? WHERE user_id = ?;`, newUsername, userID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Update failed")
		return fmt.Errorf("can't update the username: %w", err)
	}

	// Check if there was a user with the given user ID
	affectedRows, err := res.RowsAffected()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to check if the username was updated")
		return fmt.Errorf("can't check if the username was updated: %w", err)
	} else if affectedRows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// GetUsername queries the database for the username of the user with the given user ID. If the user is not found, it returns an errorUserNotFound.
func (db *AppDatabase) GetUsername(ctx context.Context, userId int64) (string, error) {

	ctx, span := tracer.Start(ctx, "database.GetUsername")
	defer span.End()

	var username string

	err := db.c.QueryRowContext(ctx, `SELECT username FROM users WHERE user_id = ?;`, userId).Scan(&username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUserNotFound
	} else if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return "", fmt.Errorf("can't get the username: %w", err)
	}

	return username, nil
}

// GetUserProfileStats returns the number of uploads, followers, and following of the user with the given user ID.
func (db *AppDatabase) GetUserProfileStats(ctx context.Context, userId int64) (int64, int64, int64, error) {

	ctx, span := tracer.Start(ctx, "database.GetUserProfileStats")
	defer span.End()

	var uploads, followers, following int64

	err := db.c.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM photos WHERE user_id = ?) AS uploads,
															 (SELECT COUNT(*) FROM follows WHERE followed_user = ?) AS followers,
															 (SELECT COUNT(*) FROM follows WHERE user_id = ?) AS following;`,
		userId, userId, userId).Scan(&uploads, &followers, &following)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return 0, 0, 0, fmt.Errorf("can't get the information: %w", err)
	}

	return uploads, followers, following, nil
}

// GetUserStream returns a slice with the photos of the given user stream. The stream is composed of the photos of the users
// that the given user follows ordered by date.
func (db *AppDatabase) GetUserStream(ctx context.Context, userId int64) ([]Photo, error) {

	ctx, span := tracer.Start(ctx, "database.GetUserStream")
	defer span.End()

	rows, err := db.c.QueryContext(ctx, `SELECT * FROM photos WHERE user_id IN (SELECT followed_user FROM follows WHERE user_id = ?) ORDER BY date DESC;`, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("can't get the photos: %w", err)
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		if err := rows.Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to scan the photos")
			return nil, fmt.Errorf("can't scan the photos: %w", err)
		}
		var photo Photo
		if err := rows.Scan(&photo.UserId, &photo.PhotoId, &photo.Path, &photo.Date); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to scan the photo")
			return nil, fmt.Errorf("can't scan the photo: %w", err)
		}
		photos = append(photos, photo)
	}

	return photos, nil
}
