package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"image"

	"github.com/aleiis/WASAPhoto/service/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type AppDatabaseI interface {
	GetUserId(ctx context.Context, username string) (int64, error)
	CreateUser(ctx context.Context, username string) (int64, error)
	UserExists(ctx context.Context, userId int64) (bool, error)
	SetUsername(ctx context.Context, userID int64, newUsername string) error
	GetUsername(ctx context.Context, userId int64) (string, error)
	GetUserProfileStats(ctx context.Context, userId int64) (int64, int64, int64, error)
	GetUserStream(ctx context.Context, userId int64) ([]Photo, error)

	PhotoExists(ctx context.Context, userId int64, photoId int64) (bool, error)
	UploadPhoto(ctx context.Context, userId int64, img image.Image, format string) error
	DeletePhoto(ctx context.Context, userId int64, photoId int64) error
	GetPhoto(ctx context.Context, userId int64, photoId int64) (Photo, error)
	GetUserPhotos(ctx context.Context, userId int64) ([]Photo, error)
	GetPhotoStats(ctx context.Context, userId int64, photoId int64) (int64, int64, error)
	GetPhotoAbsolutePath(ctx context.Context, userId int64, photoId int64) (string, error)
	GetMostRecentPhoto(ctx context.Context, userId int64) (Photo, error)

	FollowExists(ctx context.Context, userId int64, followUserId int64) (bool, error)
	CreateFollow(ctx context.Context, userId int64, followUserId int64) error
	DeleteFollow(ctx context.Context, userId int64, followUserId int64) error

	BanExists(ctx context.Context, userId int64, bannedUserId int64) (bool, error)
	CreateBan(ctx context.Context, userId int64, bannedUserId int64) error
	DeleteBan(ctx context.Context, userId int64, bannedUserId int64) error

	LikeExists(ctx context.Context, ownerId int64, photoId int64, userId int64) (bool, error)
	CreateLike(ctx context.Context, ownerId int64, photoId int64, userId int64) error
	DeleteLike(ctx context.Context, ownerId int64, photoId int64, userId int64) error

	CommentExists(ctx context.Context, photoOwner int64, photoId int64, commentId int64) (bool, error)
	CreateComment(ctx context.Context, photoOwner int64, photoId int64, commentOwner int64, content string) (int64, error)
	DeleteComment(ctx context.Context, photoOwner int64, photoId int64, commentId int64) error
	GetCommentOwner(ctx context.Context, photoOwner int64, photoId int64, commentId int64) (int64, error)
	GetPhotoComments(ctx context.Context, photoOwner int64, photoId int64) ([]Comment, error)

	Ping() error
}

type AppDatabase struct {
	c   *sql.DB
	dsn string
}

var tracer trace.Tracer = otel.Tracer("WASAPhoto/service/database")

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

	cfg, _ := config.GetConfig()

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

	if cfg.DB.MySQLExporter.Enabled {
		stmt := fmt.Sprintf("CREATE USER '%s'@'%s' IDENTIFIED BY '%s' WITH MAX_USER_CONNECTIONS 3;", cfg.DB.MySQLExporter.User, cfg.DB.MySQLExporter.Address, cfg.DB.MySQLExporter.Password)
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}

		stmt = fmt.Sprintf("GRANT PROCESS, REPLICATION CLIENT, SELECT ON *.* TO '%s'@'%s';", cfg.DB.MySQLExporter.User, cfg.DB.MySQLExporter.Address)
		_, err = db.Exec(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *AppDatabase) Ping() error {
	return db.c.Ping()
}
