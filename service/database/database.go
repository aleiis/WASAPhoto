/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/

package database

import (
	"database/sql"
	"errors"
	"fmt"
	"image"
)

type AppDatabaseI interface {
	GetUserId(username string) (int64, error)
	CreateUser(username string) (int64, error)
	UserExists(userId int64) (bool, error)
	SetUsername(userID int64, newUsername string) error
	GetUsername(userId int64) (string, error)
	GetUserProfileStats(userId int64) (int64, int64, int64, error)
	GetUserStream(userId int64) ([]Photo, error)

	PhotoExists(userId int64, photoId int64) (bool, error)
	UploadPhoto(userId int64, img image.Image, format string) error
	DeletePhoto(userId int64, photoId int64) error
	GetPhoto(userId int64, photoId int64) (Photo, error)
	GetUserPhotos(userId int64) ([]Photo, error)
	GetPhotoStats(userId int64, photoId int64) (int64, int64, error)
	GetPhotoAbsolutePath(userId int64, photoId int64) (string, error)
	GetMostRecentPhoto(userId int64) (Photo, error)

	FollowExists(userId int64, followUserId int64) (bool, error)
	CreateFollow(userId int64, followUserId int64) error
	DeleteFollow(userId int64, followUserId int64) error

	BanExists(userId int64, bannedUserId int64) (bool, error)
	CreateBan(userId int64, bannedUserId int64) error
	DeleteBan(userId int64, bannedUserId int64) error

	LikeExists(ownerId int64, photoId int64, userId int64) (bool, error)
	CreateLike(ownerId int64, photoId int64, userId int64) error
	DeleteLike(ownerId int64, photoId int64, userId int64) error

	CommentExists(photoOwner int64, photoId int64, commentId int64) (bool, error)
	CreateComment(photoOwner int64, photoId int64, commentOwner int64, content string) (int64, error)
	DeleteComment(photoOwner int64, photoId int64, commentId int64) error
	GetCommentOwner(photoOwner int64, photoId int64, commentId int64) (int64, error)
	GetPhotoComments(photoOwner int64, photoId int64) ([]Comment, error)

	Ping() error
}

type AppDatabase struct {
	c   *sql.DB
	dsn string
}

// New creates a new AppDatabase instance which is a wrapper around the provided database connection that implements the
// AppDatabaseI interface.
func New(db *sql.DB, dsn string) (AppDatabaseI, error) {
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
		c:   db,
		dsn: dsn,
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
		return err
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
		return err
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
		return err
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
		return err
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
		return err
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
		return err
	}

	return nil
}

func (db *AppDatabase) Ping() error {
	return db.c.Ping()
}
