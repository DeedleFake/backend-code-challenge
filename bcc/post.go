package bcc

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Post mirrors a row of the posts table.
type Post struct {
	ID        uint64    `db:"id" json:"id"`
	Title     string    `db:"title" json:"title"`
	Body      string    `db:"body" json:"body"`
	UserID    uint64    `db:"user_id" json:"user_id"`
	PostedAt  time.Time `db:"posted_at" json:"posted_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// GetPostByID retrieves a post from the database by its ID.
func GetPostByID(db *sqlx.DB, id uint64) (Post, error) {
	row := db.QueryRowx(`SELECT * FROM posts WHERE id=$1`, id)

	var post Post
	err := row.StructScan(&post)
	return post, err
}

// CreatePost creates a post and adds it to the database.
func CreatePost(db *sqlx.DB, userID uint64, title, body string) error {
	_, err := db.Exec(`
		INSERT INTO posts (
			user_id,
			title,
			body
		) VALUES ($1, $2, $3)
	`, userID, title, body)
	return err
}

// Comment mirrors a row of the comments table.
type Comment struct {
	ID          uint64    `db:"id" json:"id"`
	UserID      uint64    `db:"user_id" json:"user_id"`
	PostID      uint64    `db:"post_id" json:"post_id"`
	Message     string    `db:"message" json:"message"`
	CommentedAt time.Time `db:"commented_at" json:"commented_at"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// GetCommentsByPostID returns an iterator of Comments on a given
// post, sorted in ascending post time order.
func GetCommentsByPostID(db *sqlx.DB, postID uint64) (*Iterator, error) {
	rows, err := db.Queryx(`SELECT * FROM comments WHERE post_id=$1`, postID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	return &Iterator{
		next: rows.Next,
		cur: func() (interface{}, error) {
			var comment Comment
			err := rows.StructScan(&comment)
			return comment, err
		},
		close: rows.Close,
	}, nil
}

// CreateComment creates a comment on a post.
func CreateComment(db *sqlx.DB, userID, postID uint64, message string) error {
	_, err := db.Exec(`
		INSERT INTO comments (
			user_id,
			post_id,
			message
		) VALUES ($1, $2, $3)
	`, userID, postID, message)
	return err
}

// DeleteComment deletes a comment.
func DeleteComment(db *sqlx.DB, commentID uint64) error {
	_, err := db.Exec(`DELETE FROM comments WHERE id = $1`, commentID)
	return err
}
