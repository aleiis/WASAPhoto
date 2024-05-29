package database

import (
	"fmt"
)

type Comment struct {
	PhotoOwner   int64
	PhotoId      int64
	CommentId    int64
	CommentOwner int64
	Content      string
}

func (db *AppDatabase) CommentExists(photoOwner int64, photoId int64, commentId int64) (bool, error) {
	var count int
	err := db.c.QueryRow(`SELECT COUNT(*) FROM comments WHERE photo_owner = ? AND photo_id = ? AND comment_id = ?;`, photoOwner, photoId, commentId).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("can't check if the comment exists: %w", err)
	}

	return count > 0, nil
}

func (db *AppDatabase) CreateComment(photoOwner int64, photoId int64, commentOwner int64, content string) (int64, error) {
	if len(content) == 0 || len(content) > 128 {
		return -1, fmt.Errorf("the content must measure between 1 and 128 bytes")
	}

	// Calculate the comment ID
	var count int64
	err := db.c.QueryRow(`SELECT COUNT(*) FROM comments WHERE photo_owner = ? AND photo_id = ?;`, photoOwner, photoId).Scan(&count)
	if err != nil {
		return -1, fmt.Errorf("can't get the number of comments: %w", err)
	}

	// Insert the comment
	_, err = db.c.Exec(`INSERT INTO comments (photo_owner, photo_id, comment_id, comment_owner, content) VALUES (?, ?, ?, ?, ?);`, photoOwner, photoId, count, commentOwner, content)
	if err != nil {
		return -1, fmt.Errorf("can't insert the comment: %w", err)
	}

	return count, nil
}

func (db *AppDatabase) DeleteComment(photoOwner int64, photoId int64, commentId int64) error {
	_, err := db.c.Exec(`DELETE FROM comments WHERE photo_owner = ? AND photo_id = ? AND comment_id = ?;`, photoOwner, photoId, commentId)
	if err != nil {
		return fmt.Errorf("can't delete the comment: %w", err)
	}

	_, err = db.c.Exec(`UPDATE comments SET comment_id = comment_id - 1 WHERE photo_owner = ? AND photo_id = ? AND comment_id > ?;`, photoOwner, photoId, commentId)
	if err != nil {
		return fmt.Errorf("can't update ids after comment delete (DB CORRUPTION!!!): %w", err)
	}

	return nil
}

func (db *AppDatabase) GetCommentOwner(photoOwner int64, photoId int64, commentId int64) (int64, error) {
	var commentOwner int64
	err := db.c.QueryRow(`SELECT comment_owner FROM comments WHERE photo_owner = ? AND photo_id = ? AND comment_id = ?;`, photoOwner, photoId, commentId).Scan(&commentOwner)
	if err != nil {
		return 0, fmt.Errorf("can't get the comment owner: %w", err)
	}

	return commentOwner, nil
}

func (db *AppDatabase) GetPhotoComments(photoOwner int64, photoId int64) ([]Comment, error) {

	// Get the comments of the photo
	rows, err := db.c.Query(`SELECT * FROM comments WHERE photo_owner = ? AND photo_id = ?;`, photoOwner, photoId)
	if err != nil {
		return nil, fmt.Errorf("can't get the comments of the photo: %w", err)
	}
	defer rows.Close()

	// Scan the photos from the query result
	var comments []Comment
	for rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("can't scan the comments: %w", err)
		}
		var comment Comment
		if err := rows.Scan(&comment.PhotoOwner, &comment.PhotoId, &comment.CommentId, &comment.CommentOwner, &comment.Content); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, nil
}
