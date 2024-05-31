package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

var ErrPhotoNotFound = errors.New("photo not found")
var ErrUnsupportedImageFormat = errors.New("unsupported image format")

type Photo struct {
	UserId  int64
	PhotoId int64
	Path    string
	Date    string
}

func (db *AppDatabase) PhotoExists(userId int64, photoId int64) (bool, error) {
	var count int
	err := db.c.QueryRow(`SELECT COUNT(*) FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("can't get the number of photos: %w", err)
	}

	return count > 0, nil
}

func (db *AppDatabase) UploadPhoto(userId int64, img image.Image, format string) error {

	// Check if the user exists
	exists, err := db.UserExists(userId)
	if err != nil {
		return fmt.Errorf("can't check if the user exists: %w", err)
	} else if !exists {
		return ErrUserNotFound
	}

	// Get the number of photos of the user
	var count int
	err = db.c.QueryRow(`SELECT COUNT(*) FROM photos WHERE user_id = ?;`, userId).Scan(&count)
	if err != nil {
		return fmt.Errorf("can't get the number of photos: %w", err)
	}

	photoFilename := fmt.Sprintf("%s.%s", uuid.New().String(), format)

	// Check if all the folder structure exists
	photoPath := filepath.Join(filepath.Dir(db.dsn), "images", fmt.Sprint(userId))
	if _, err := os.Stat(photoPath); os.IsNotExist(err) {
		if err := os.MkdirAll(photoPath, 0755); err != nil {
			return fmt.Errorf("can't create the folder structure for user with id %d: %w", userId, err)
		}
	}
	photoPath = filepath.Join(photoPath, photoFilename)

	// Start a transaction
	tx, err := db.c.Begin()
	if err != nil {
		return fmt.Errorf("can't begin transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			return
		}
	}(tx)

	// Insert the photo data
	relativePath := filepath.Join(fmt.Sprint(userId), photoFilename)
	relativePath = filepath.ToSlash(relativePath)
	_, err = tx.Exec(`INSERT INTO photos (user_id, photo_id, path, date) VALUES (?, ?, ?, datetime('now'));`, userId, count, relativePath)
	if err != nil {
		return fmt.Errorf("can't insert photo data: %w", err)
	}

	// Save the photo
	f, err := os.Create(photoPath)
	if err != nil {
		return fmt.Errorf("can't create the file: %w", err)
	}

	switch format {
	case "jpg", "jpeg":
		err = jpeg.Encode(f, img, nil)
		if err != nil {
			_ = f.Close()
			_ = os.Remove(photoPath)
			return fmt.Errorf("can't encode the image: %w", err)

		}
	case "png":
		err = png.Encode(f, img)
		if err != nil {
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
		_ = os.Remove(photoPath)
		return fmt.Errorf("can't commit transaction: %w", err)
	}

	return nil
}

func (db *AppDatabase) DeletePhoto(userId int64, photoId int64) error {

	// Check if the photo exists
	if exists, err := db.PhotoExists(userId, photoId); err != nil {
		return fmt.Errorf("can't check if the photo exists: %w", err)
	} else if !exists {
		return ErrPhotoNotFound
	}

	// Get the photo path
	var photoPath string
	err := db.c.QueryRow(`SELECT path FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId).Scan(&photoPath)
	if err != nil {
		return fmt.Errorf("can't get the photo path: %w", err)
	}

	photoPath = filepath.FromSlash(photoPath)
	photoPath = filepath.Join(filepath.Dir(db.dsn), "images", photoPath)

	// Start a transaction
	tx, err := db.c.Begin()
	if err != nil {
		return fmt.Errorf("can't begin transaction: %w", err)
	}
	defer func(tx *sql.Tx) {
		err := tx.Rollback()
		if err != nil {
			return
		}
	}(tx)

	// Delete the photo from the database
	_, err = tx.Exec(`DELETE FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId)
	if err != nil {
		return fmt.Errorf("can't delete the photo: %w", err)
	}

	// Delete the photo from the filesystem
	if err := os.Remove(photoPath); err != nil {
		return fmt.Errorf("can't delete the photo from the filesystem: %w", err)
	}

	// Update the photo IDs
	_, err = tx.Exec(`UPDATE photos SET photo_id = photo_id - 1 WHERE user_id = ? AND photo_id > ?;`, userId, photoId)
	if err != nil {
		return fmt.Errorf("can't update the photo IDs: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("can't commit transaction: %w", err)
	}

	return nil
}

func (db *AppDatabase) GetPhoto(userId int64, photoId int64) (Photo, error) {

	var photo Photo
	err := db.c.QueryRow(`SELECT * FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId).Scan(&photo.UserId, &photo.PhotoId, &photo.Path, &photo.Date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Photo{}, ErrPhotoNotFound
		} else {
			return Photo{}, err
		}
	}

	return photo, nil

}

// GetUserPhotos returns the photos of the user with the given user ID.
func (db *AppDatabase) GetUserPhotos(userId int64) ([]Photo, error) {

	// Check if the user exists
	if exists, err := db.UserExists(userId); err != nil {
		return nil, fmt.Errorf("can't check if the user exists: %w", err)
	} else if !exists {
		return nil, ErrUserNotFound
	}

	// Get the photos of the user
	rows, err := db.c.Query(`SELECT * FROM photos WHERE user_id = ? ORDER BY date DESC;`, userId)
	if err != nil {
		return nil, fmt.Errorf("can't get the photos of the user: %w", err)
	}
	defer rows.Close()

	// Scan the photos from the query result
	var photos []Photo
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("can't scan the photos: %w", err)
		}
		var photo Photo
		if err := rows.Scan(&photo.UserId, &photo.PhotoId, &photo.Path, &photo.Date); err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}

	return photos, nil
}

func (db *AppDatabase) GetPhotoStats(userId int64, photoId int64) (int64, int64, error) {

	// Get the number of likes and comments of the photo
	var likes, comments int64
	err := db.c.QueryRow(`SELECT (SELECT COUNT(*) FROM likes WHERE photo_owner = ? AND photo_id = ?) AS likes,
															 (SELECT COUNT(*) FROM comments WHERE photo_owner = ? AND photo_id = ?) AS comments;`,
		userId, photoId, userId, photoId).Scan(&likes, &comments)
	if err != nil {
		return 0, 0, err
	}

	return likes, comments, nil
}

func (db *AppDatabase) GetPhotoAbsolutePath(userId int64, photoId int64) (string, error) {

	var photoPath string
	err := db.c.QueryRow(`SELECT path FROM photos WHERE user_id = ? AND photo_id = ?;`, userId, photoId).Scan(&photoPath)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrPhotoNotFound
		} else {
			return "", err
		}
	}

	photoPath = filepath.FromSlash(photoPath)
	photoPath = filepath.Join(filepath.Dir(db.dsn), "images", photoPath)
	if filepath.IsAbs(photoPath) {
		return photoPath, nil
	}

	photoPath, err = filepath.Abs(photoPath)
	if err != nil {
		return "", fmt.Errorf("can't get the absolute path: %w", err)
	}

	return photoPath, nil
}

func (db *AppDatabase) GetMostRecentPhoto(userId int64) (Photo, error) {

	var photo Photo
	err := db.c.QueryRow(`SELECT * FROM photos WHERE user_id = ? ORDER BY date DESC LIMIT 1;`, userId).Scan(&photo.UserId, &photo.PhotoId, &photo.Path, &photo.Date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Photo{}, errors.New("no photos found")
		} else {
			return Photo{}, err
		}
	}

	return photo, nil
}
