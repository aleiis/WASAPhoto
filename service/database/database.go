package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
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
	err := db.QueryRow(`SELECT TABLE_NAME FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users';`).Scan(&tableName)
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
				user_id INTEGER PRIMARY KEY AUTO_INCREMENT,
				username VARCHAR(16) UNIQUE NOT NULL COLLATE utf8_general_ci
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

func logDatabase(db *sql.DB, db_logger *logrus.Logger) {
	_, err := db.Exec("CREATE VIRTUAL TABLE IF NOT EXISTS temp.stat USING dbstat(main)")
	if err != nil {
		db_logger.WithError(err).Errorf("Error creating virtual table")
		return
	}

	query := `SELECT name, pageno, pagetype, ncell, payload, unused, mx_payload, pgoffset, pgsize 
              FROM dbstat`

	rows, err := db.Query(query)
	if err != nil {
		db_logger.WithError(err).Errorf("Error executing dbstat query")
		return
	}
	defer rows.Close()

	for rows.Next() {
		var (
			name, pagetype                                              sql.NullString
			pageno, ncell, payload, unused, mxPayload, pgOffset, pgSize sql.NullInt64
		)

		err := rows.Scan(&name, &pageno, &pagetype, &ncell, &payload, &unused, &mxPayload, &pgOffset, &pgSize)
		if err != nil {
			db_logger.WithError(err).Errorf("Error scanning dbstat row")
			continue
		}

		db_logger.WithFields(logrus.Fields{
			"name":      name.String,
			"pageno":    pageno.Int64,
			"pagetype":  pagetype.String,
			"ncell":     ncell.Int64,
			"payload":   payload.Int64,
			"unused":    unused.Int64,
			"mxPayload": mxPayload.Int64,
			"pgOffset":  pgOffset.Int64,
			"pgSize":    pgSize.Int64,
		}).Info("Logging dbstat entry")
	}

	if err := rows.Err(); err != nil {
		db_logger.WithError(err).Errorf("Error iterating dbstat rows")
	}
}

func StartPeriodicLogging(ctx context.Context, db *sql.DB, interval time.Duration, logFile string) {
	db_logger := logrus.New()

	if logFile != "" {
		if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
			db_logger.WithError(err).Error("can't create the log directory")
			return
		}
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			db_logger.WithError(err).Error("can't open the log file")
			return
		}
		defer file.Close()
		db_logger.SetOutput(file)
	} else {
		db_logger.SetOutput(os.Stdout)
	}

	db_logger.SetLevel(logrus.InfoLevel)

	db_logger.Info("Starting the database logging")

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			db_logger.Info("Stopping database logging...")
			return
		case <-ticker.C:
			logDatabase(db, db_logger)
		}
	}
}
