package database

import (
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUsernameAlreadyExists = errors.New("username already exists")

// It queries the database for the user ID of the given username. If the user is not found, it returns an errorUserNotFound.
func (db *AppDatabase) GetUserId(username string) (int64, error) {
	var id int64

	err := db.c.QueryRow(`SELECT user_id FROM users WHERE username = ?;`, username).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return -1, ErrUserNotFound
	} else if err != nil {
		return -1, fmt.Errorf("can't get the user ID: %w", err)
	}

	return id, nil
}

// It creates a new user with the given username. It returns the user ID of the new user.
// If the username already exists, it returns an errUsernameAlreadyExists.
func (db *AppDatabase) CreateUser(username string) (int64, error) {

	// Insert the new user into the database
	res, err := db.c.Exec(`INSERT INTO users(username) VALUES (?);`, username)
	if err != nil {
		if _, err := db.GetUserId(username); err == nil {
			return -1, ErrUsernameAlreadyExists
		} else {
			return -1, fmt.Errorf("can't create the user: %w", err)
		}
	}

	// Get the user ID of the new user
	id, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("can't get the user ID: %w", err)
	}

	return id, nil
}

// It checks if the user ID exists in the database.
func (db *AppDatabase) UserIdExists(userId int64) (bool, error) {
	var id int64

	err := db.c.QueryRow(`SELECT user_id FROM users WHERE user_id = ?;`, userId).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("can't check if the user ID exists: %w", err)
	}

	return true, nil
}

// It sets the username of the user with the given user ID.
func (db *AppDatabase) SetUsername(userID int64, newUsername string) error {

	// Check if the new username already exists
	_, err := db.GetUserId(newUsername)
	if err != ErrUserNotFound {
		if err == nil {
			return ErrUsernameAlreadyExists
		} else {
			return fmt.Errorf("can't check if the username already exists: %w", err)
		}
	}

	// Update the username of the user
	res, err := db.c.Exec(`UPDATE users SET username = ? WHERE user_id = ?;`, newUsername, userID)
	if err != nil {
		return fmt.Errorf("can't update the username: %w", err)
	}

	// Check if there was a user with the given user ID
	affectedRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("can't check if the username was updated: %w", err)
	} else if affectedRows == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (db *AppDatabase) GetUsername(userId int64) (string, error) {
	var username string

	err := db.c.QueryRow(`SELECT username FROM users WHERE user_id = ?;`, userId).Scan(&username)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrUserNotFound
	} else if err != nil {
		return "", fmt.Errorf("can't get the username: %w", err)
	}

	return username, nil
}

// It returns the number of uploads, followers, and following of the user with the given user ID.
func (db *AppDatabase) GetUserProfileStats(userId int64) (int64, int64, int64, error) {
	var uploads, followers, following int64

	err := db.c.QueryRow(`SELECT (SELECT COUNT(*) FROM photos WHERE user_id = ?) AS uploads,
															 (SELECT COUNT(*) FROM follows WHERE followed_user = ?) AS followers,
															 (SELECT COUNT(*) FROM follows WHERE user_id = ?) AS following;`,
		userId, userId, userId).Scan(&uploads, &followers, &following)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("can't get the information: %w", err)
	}

	return uploads, followers, following, nil
}

func (db *AppDatabase) GetUserStream(userId int64) ([]Photo, error) {
	rows, err := db.c.Query(`SELECT * FROM photos WHERE user_id IN (SELECT followed_user FROM follows WHERE user_id = ?) ORDER BY date DESC;`, userId)
	if err != nil {
		return nil, fmt.Errorf("can't get the photos: %w", err)
	}
	defer rows.Close()

	var photos []Photo
	for rows.Next() {
		var photo Photo
		if err := rows.Scan(&photo.UserId, &photo.PhotoId, &photo.Path, &photo.Date); err != nil {
			return nil, fmt.Errorf("can't scan the photo: %w", err)
		}
		photos = append(photos, photo)
	}

	return photos, nil
}
