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

func (db *AppDatabase) CreateBan(userId int64, bannedUserId int64) error {

	// Check if the user is trying to follow itself
	if userId == bannedUserId {
		return ErrBanYourself
	}

	// Insert the follow
	_, err := db.c.Exec(`INSERT INTO bans (user_id, banned_user) VALUES (?, ?);`, userId, bannedUserId)
	if err != nil {
		return fmt.Errorf("db insert error: %w", err)
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
