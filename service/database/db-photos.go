package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/aleiis/WASAPhoto/service/config"
	"go.opentelemetry.io/otel/codes"

	"github.com/google/uuid"
)

var ErrPhotoNotFound = errors.New("photo not found")
var ErrUnsupportedImageFormat = errors.New("unsupported image format")

type Photo struct {
	UserId  int64
	PhotoId int64
	Path    string
	Date    string
}

func (db *AppDatabase) PhotoExists(ctx context.Context, userId int64, photoId int64) (bool, error) {

	ctx, span := tracer.Start(ctx, "database.PhotoExists")
	defer span.End()

	var count int
	err := db.c.QueryRowContext(ctx, `SELECT COUNT(*) FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("can't get the number of photos: %w", err)
	}

	return count > 0, nil
}

func (db *AppDatabase) UploadPhoto(ctx context.Context, userId int64, img image.Image, format string) error {

	ctx, span := tracer.Start(ctx, "database.UploadPhoto")
	defer span.End()

	cfg, _ := config.GetConfig()

	// Check if the user exists
	exists, err := db.UserExists(ctx, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Can't check if the user exists")
		return fmt.Errorf("can't check if the user exists: %w", err)
	} else if !exists {
		return ErrUserNotFound
	}

	// Get the number of photos of the user
	var count int
	err = db.c.QueryRowContext(ctx, `SELECT COUNT(*) FROM photos WHERE user_id = ?;`, userId).Scan(&count)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return fmt.Errorf("can't get the number of photos: %w", err)
	}

	photoFilename := fmt.Sprintf("%s.%s", uuid.New().String(), format)

	// Check if all the folder structure exists
	photoPath := filepath.Join(cfg.ImageStorage.Path, fmt.Sprint(userId))
	if _, err := os.Stat(photoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(photoPath, 0755); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to create folder structure")
			return fmt.Errorf("can't create the folder structure for user with id %d: %w", userId, err)
		}
	}
	photoPath = filepath.Join(photoPath, photoFilename)

	// Start a transaction
	tx, err := db.c.Begin()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to start transaction")
		return fmt.Errorf("can't begin transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// Insert the photo data
	relativePath := filepath.Join(fmt.Sprint(userId), photoFilename)
	relativePath = filepath.ToSlash(relativePath)
	_, err = tx.ExecContext(ctx, `INSERT INTO photos (user_id, photo_id, path, date) VALUES (?, ?, ?, NOW());`, userId, count, relativePath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to insert photo data")
		return fmt.Errorf("can't insert photo data: %w", err)
	}

	// Save the photo
	f, err := os.Create(photoPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to create file")
		return fmt.Errorf("can't create the file: %w", err)
	}

	switch format {
	case "jpg", "jpeg":
		err = jpeg.Encode(f, img, nil)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to encode image")
			_ = f.Close()
			_ = os.Remove(photoPath)
			return fmt.Errorf("can't encode the image: %w", err)

		}
	case "png":
		err = png.Encode(f, img)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed to encode image")
			_ = f.Close()
			_ = os.Remove(photoPath)
			return fmt.Errorf("can't encode the image: %w", err)

		}
	default:
		_ = f.Close()
		_ = os.Remove(photoPath)
		return ErrUnsupportedImageFormat
	}

	_ = f.Close()
	err = tx.Commit()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to commit transaction")
		_ = os.Remove(photoPath)
		return fmt.Errorf("can't commit transaction: %w", err)
	}

	return nil
}

func (db *AppDatabase) DeletePhoto(ctx context.Context, userId int64, photoId int64) error {

	ctx, span := tracer.Start(ctx, "database.DeletePhoto")
	defer span.End()

	cfg, _ := config.GetConfig()

	// Check if the photo exists
	if exists, err := db.PhotoExists(ctx, userId, photoId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Can't check if the photo exists")
		return fmt.Errorf("can't check if the photo exists: %w", err)
	} else if !exists {
		return ErrPhotoNotFound
	}

	// Get the photo path
	var photoPath string
	err := db.c.QueryRowContext(ctx, `SELECT path FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId).Scan(&photoPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get the photo path")
		return fmt.Errorf("can't get the photo path: %w", err)
	}

	photoPath = filepath.FromSlash(photoPath)
	photoPath = filepath.Join(cfg.ImageStorage.Path, photoPath)

	// Start a transaction
	tx, err := db.c.Begin()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to start transaction")
		return fmt.Errorf("can't begin transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	// Delete the photo from the database
	_, err = tx.ExecContext(ctx, `DELETE FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete photo")
		return fmt.Errorf("can't delete the photo: %w", err)
	}

	// Delete the photo from the filesystem
	if err := os.Remove(photoPath); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to delete photo from filesystem")
		return fmt.Errorf("can't delete the photo from the filesystem: %w", err)
	}

	// Update the photo IDs
	_, err = tx.ExecContext(ctx, `UPDATE photos SET photo_id = photo_id - 1 WHERE user_id = ? AND photo_id > ?;`, userId, photoId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to update photo IDs")
		return fmt.Errorf("can't update the photo IDs: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to commit transaction")
		return fmt.Errorf("can't commit transaction: %w", err)
	}

	return nil
}

func (db *AppDatabase) GetPhoto(ctx context.Context, userId int64, photoId int64) (Photo, error) {

	ctx, span := tracer.Start(ctx, "database.GetPhoto")
	defer span.End()

	var photo Photo
	err := db.c.QueryRowContext(ctx, `SELECT * FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId).Scan(&photo.UserId, &photo.PhotoId, &photo.Path, &photo.Date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Photo{}, ErrPhotoNotFound
		} else {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Query failed")
			return Photo{}, err
		}
	}

	return photo, nil

}

// GetUserPhotos returns the photos of the user with the given user ID.
func (db *AppDatabase) GetUserPhotos(ctx context.Context, userId int64) ([]Photo, error) {

	ctx, span := tracer.Start(ctx, "database.GetUserPhotos")
	defer span.End()

	// Check if the user exists
	if exists, err := db.UserExists(ctx, userId); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Can't check if the user exists")
		return nil, fmt.Errorf("can't check if the user exists: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	// Get the photos of the user
	rows, err := db.c.QueryContext(ctx, `SELECT * FROM photos WHERE user_id = ? ORDER BY date DESC;`, userId)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("can't get the photos of the user: %w", err)
	}
	defer rows.Close()

	// Scan the photos from the query result
	var photos []Photo
	for rows.Next() {
		if err := rows.Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Failed iterating rows")
			return nil, fmt.Errorf("can't scan the photos: %w", err)
		}
		var photo Photo
		if err := rows.Scan(&photo.UserId, &photo.PhotoId, &photo.Path, &photo.Date); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Scan failed")
			return nil, err
		}
		photos = append(photos, photo)
	}

	return photos, nil
}

func (db *AppDatabase) GetPhotoStats(ctx context.Context, userId int64, photoId int64) (int64, int64, error) {

	ctx, span := tracer.Start(ctx, "database.GetPhotoStats")
	defer span.End()

	// Get the number of likes and comments of the photo
	var likes, comments int64
	err := db.c.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM likes WHERE photo_owner = ? AND photo_id = ?) AS likes,
															 (SELECT COUNT(*) FROM comments WHERE photo_owner = ? AND photo_id = ?) AS comments;`,
		userId, photoId, userId, photoId).Scan(&likes, &comments)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return 0, 0, err
	}

	return likes, comments, nil
}

func (db *AppDatabase) GetPhotoAbsolutePath(ctx context.Context, userId int64, photoId int64) (string, error) {

	ctx, span := tracer.Start(ctx, "database.GetPhotoAbsolutePath")
	defer span.End()

	cfg, _ := config.GetConfig()

	var photoPath string
	err := db.c.QueryRowContext(ctx, `SELECT path FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId).Scan(&photoPath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrPhotoNotFound
		} else {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Query failed")
			return "", err
		}
	}

	photoPath = filepath.FromSlash(photoPath)
	photoPath = filepath.Join(cfg.ImageStorage.Path, photoPath)
	if filepath.IsAbs(photoPath) {
		return photoPath, nil
	}

	photoPath, err = filepath.Abs(photoPath)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to get the absolute path")
		return "", fmt.Errorf("can't get the absolute path: %w", err)
	}

	return photoPath, nil
}

func (db *AppDatabase) GetMostRecentPhoto(ctx context.Context, userId int64) (Photo, error) {

	ctx, span := tracer.Start(ctx, "database.GetMostRecentPhoto")
	defer span.End()

	var photo Photo
	err := db.c.QueryRowContext(ctx, `SELECT * FROM photos WHERE user_id = ? ORDER BY date DESC LIMIT 1;`, userId).Scan(&photo.UserId, &photo.PhotoId, &photo.Path, &photo.Date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Photo{}, errors.New("no photos found")
		} else {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Query failed")
			return Photo{}, err
		}
	}

	return photo, nil
}
