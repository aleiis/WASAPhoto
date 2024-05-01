package database

import (
	"fmt"
)

func (db *AppDatabase) LikeExists(ownerId int64, photoId int64, userId int64) (bool, error) {
	var count int
	err := db.c.QueryRow(`SELECT COUNT(*) FROM likes WHERE photo_owner = ? AND photo_id = ? AND user_id = ?;`, ownerId, photoId, userId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("can't check if the like exists: %w", err)
	}

	return count > 0, nil
}

func (db *AppDatabase) CreateLike(ownerId int64, photoId int64, userId int64) error {
	_, err := db.c.Exec(`INSERT INTO likes (photo_owner, photo_id, user_id) VALUES (?, ?, ?);`, ownerId, photoId, userId)
	if err != nil {
		return fmt.Errorf("can't insert the like: %w", err)
	}

	return nil
}

func (db *AppDatabase) DeleteLike(ownerId int64, photoId int64, userId int64) error {
	_, err := db.c.Exec(`DELETE FROM likes WHERE photo_owner = ? AND photo_id = ? AND user_id = ?;`, ownerId, photoId, userId)
	if err != nil {
		return fmt.Errorf("can't delete the like: %w", err)
	}

	return nil
}
