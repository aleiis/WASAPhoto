package database

import (
	"errors"
	"fmt"
)

var ErrBanYourself = errors.New("can't ban yourself")

func (db *AppDatabase) BanExists(userId int64, bannedUserId int64) (bool, error) {

	var count int
	err := db.c.QueryRow(`SELECT COUNT(*) FROM bans WHERE user_id = ? AND banned_user = ?;`, userId, bannedUserId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("query error: %w", err)
	}

	return count > 0, nil
}

// CreateBan registers a ban in the database. If the banned user was following the user, the follow is deleted.
// It returns an error if the user is trying to ban itself.
func (db *AppDatabase) CreateBan(userId int64, bannedUserId int64) error {

	// Check if the user is trying to ban itself
	if userId == bannedUserId {
		return ErrBanYourself
	}

	// Create a transaction
	tx, err := db.c.Begin()
	if err != nil {
		return fmt.Errorf("can't start a transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if the banned user was following the user
	var count int
	err = tx.QueryRow(`SELECT COUNT(*) FROM follows WHERE user_id = ? AND followed_user = ?;`, bannedUserId, userId).Scan(&count)
	if err != nil {
		return fmt.Errorf("can't check if the banned user was following the user: %w", err)
	}

	// If so, delete the follow
	if count > 0 {
		_, err = tx.Exec(`DELETE FROM follows WHERE user_id = ? AND followed_user = ?;`, bannedUserId, userId)
		if err != nil {
			return fmt.Errorf("can't delete the follow: %w", err)
		}
	}

	// Insert the ban
	_, err = tx.Exec(`INSERT INTO bans (user_id, banned_user) VALUES (?, ?);`, userId, bannedUserId)
	if err != nil {
		return fmt.Errorf("db insert error: %w", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("can't commit the transaction: %w", err)
	}

	return nil
}

func (db *AppDatabase) DeleteBan(userId int64, bannedUserId int64) error {

	// Delete the follow
	_, err := db.c.Exec(`DELETE FROM bans WHERE user_id = ? AND banned_user = ?;`, userId, bannedUserId)
	if err != nil {
		return fmt.Errorf("db delete error: %w", err)
	}

	return nil
}
