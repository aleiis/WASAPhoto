package database

import (
	"errors"
	"fmt"
)

var ErrFollowYourself = errors.New("can't follow yourself")

func (db *AppDatabase) FollowExists(userId int64, followUserId int64) (bool, error) {

	var count int
	err := db.c.QueryRow(`SELECT COUNT(*) FROM follows WHERE user_id = ? AND followed_user = ?;`, userId, followUserId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}

	return count > 0, nil
}

func (db *AppDatabase) CreateFollow(userId int64, followUserId int64) error {

	// Check if the user is trying to follow itself
	if userId == followUserId {
		return ErrFollowYourself
	}

	// Insert the follow
	_, err := db.c.Exec(`INSERT INTO follows (user_id, followed_user) VALUES (?, ?);`, userId, followUserId)
	if err != nil {
		return fmt.Errorf("db insert error: %w", err)
	}

	return nil
}

func (db *AppDatabase) DeleteFollow(userId int64, followUserId int64) error {

	// Delete the follow
	_, err := db.c.Exec(`DELETE FROM follows WHERE user_id = ? AND followed_user = ?;`, userId, followUserId)
	if err != nil {
		return fmt.Errorf("db delete error: %w", err)
	}

	return nil
}
