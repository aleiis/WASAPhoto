package database

import (
	"database/sql"
	"errors"
	"fmt"
	"image"
)

type AppDatabaseI interface {

	// User functions
	GetUserId(username string) (int64, error)
	CreateUser(username string) (int64, error)
	UserIdExists(userId int64) (bool, error)
	SetUsername(userID int64, newUsername string) error
	GetUsername(userId int64) (string, error)
	GetUserProfileStats(userId int64) (int64, int64, int64, error)
	GetUserStream(userId int64) ([]Photo, error)

	// Photo functions
	PhotoExists(userId int64, photoId int64) (bool, error)
	UploadPhoto(userId int64, img image.Image, format string) error
	DeletePhoto(userId int64, photoId int64) error
	GetUserPhotos(userId int64) ([]Photo, error)
	GetPhotoStats(userId int64, photoId int64) (int64, int64, error)

	// Follow functions
	FollowExists(userId int64, followUserId int64) (bool, error)
	CreateFollow(userId int64, followUserId int64) error
	DeleteFollow(userId int64, followUserId int64) error

	// Ban functions
	BanExists(userId int64, bannedUserId int64) (bool, error)
	CreateBan(userId int64, bannedUserId int64) error
	DeleteBan(userId int64, bannedUserId int64) error

	// Like functions
	LikeExists(ownerId int64, photoId int64, userId int64) (bool, error)
	CreateLike(ownerId int64, photoId int64, userId int64) error
	DeleteLike(ownerId int64, photoId int64, userId int64) error

	// Comment functions
	CommentExists(ownerId int64, photoId int64, commentId int64) (bool, error)
	CreateComment(ownerId int64, photoId int64, commentOwner int64, content string) error
	DeleteComment(ownerId int64, photoId int64, commentId int64) error
	GetCommentOwner(ownerId int64, photoId int64, commentId int64) (int64, error)

	Ping() error
}

type AppDatabase struct {
	c *sql.DB
}

func New(db *sql.DB) (AppDatabaseI, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Check the connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Check the schema. If the schema is not present, create it.
	var tableName string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='users';`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		if err := createSchema(db); err != nil {
			return nil, fmt.Errorf("can't create the schema: %w", err)
		}
	}

	return &AppDatabase{
		c: db,
	}, nil
}

func createSchema(db *sql.DB) error {
	var err error

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS users (
				user_id INTEGER PRIMARY KEY AUTOINCREMENT,
				username VARCHAR(16) UNIQUE NOT NULL
			);
		`)
	if err != nil {
		return fmt.Errorf("can't create the schema: %w", err)
	}

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS photos (
				user_id INTEGER,
				photo_id INTEGER,
				path TEXT NOT NULL,
				date DATETIME NOT NULL,
				PRIMARY KEY (user_id, photo_id),
				FOREIGN KEY (user_id) 
					REFERENCES users(user_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE
			);
		`)
	if err != nil {
		return fmt.Errorf("can't create the schema: %w", err)
	}

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS follows (
				user_id INTEGER,
				followed_user INTEGER,
				PRIMARY KEY (user_id, followed_user),
				FOREIGN KEY (user_id) 
					REFERENCES users(user_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE,
				FOREIGN KEY (followed_user) 
					REFERENCES users(user_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE
			);
		`)
	if err != nil {
		return fmt.Errorf("can't create the schema: %w", err)
	}

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS bans (
				user_id INTEGER,
				banned_user INTEGER,
				PRIMARY KEY (user_id, banned_user),
				FOREIGN KEY (user_id) 
					REFERENCES users(user_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE,
				FOREIGN KEY (banned_user) 
					REFERENCES users(user_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE
			);
		`)
	if err != nil {
		return fmt.Errorf("can't create the schema: %w", err)
	}

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS likes (
				photo_owner INTEGER,
				photo_id INTEGER,
				user_id INTEGER,
				PRIMARY KEY (photo_owner, photo_id, user_id),
				FOREIGN KEY (photo_owner, photo_id)
					REFERENCES photos(user_id, photo_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE,
				FOREIGN KEY (user_id)
					REFERENCES users(user_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE
			);
		`)
	if err != nil {
		return fmt.Errorf("can't create the schema: %w", err)
	}

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS comments (
				photo_owner INTEGER,
				photo_id INTEGER,
				comment_id INTEGER,
				comment_owner INTEGER NOT NULL,
				content VARCHAR(128) NOT NULL,
				PRIMARY KEY (photo_owner, photo_id, comment_id),
				FOREIGN KEY (photo_owner, photo_id)
					REFERENCES photos(user_id, photo_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE,
				FOREIGN KEY (comment_owner)
					REFERENCES users(user_id)
						ON DELETE CASCADE
						ON UPDATE CASCADE
			);
		`)
	if err != nil {
		return fmt.Errorf("can't create the schema: %w", err)
	}

	return nil
}

func (db *AppDatabase) Ping() error {
	return db.c.Ping()
}
