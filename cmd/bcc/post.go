package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DeedleFake/backend-code-challenge/bcc"
	"github.com/jmoiron/sqlx"
)

func handleGetPost(rw http.ResponseWriter, req *http.Request, db *sqlx.DB) error {
	var q struct {
		PostID int `query:"post_id"`
	}
	err := parseQuery(req.URL.Query(), &q)
	if err != nil {
		return APIUserError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to parse query: %w", err),
		}
	}

	post, err := bcc.GetPostByID(db, q.PostID)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}

	comments, err := bcc.GetCommentsByPostID(db, q.PostID)
	if err != nil {
		return fmt.Errorf("comments: %w", err)
	}
	defer comments.Close()

	result := struct {
		UserID    int       `json:"user_id"`
		PostedAt  time.Time `json:"posted_at"`
		UpdatedAt time.Time `json:"updated_at"`

		Title    string        `json:"title"`
		Body     string        `json:"body"`
		Comments []interface{} `json:"comments"`
	}{
		UserID:    post.UserID,
		PostedAt:  post.PostedAt,
		UpdatedAt: post.UpdatedAt,

		Title: post.Title,
		Body:  post.Body,
	}

	for comments.Next() {
		comment := comments.Current().(bcc.Comment)

		result.Comments = append(result.Comments, struct {
			UserID    int       `json:"user_id"`
			PostedAt  time.Time `"json:"posted_at"`
			UpdatedAt time.Time `json:"updated_at"`
			ID        int       `json:"id"`
			Message   string    `json:"message"`
		}{
			UserID:    comment.UserID,
			PostedAt:  comment.CommentedAt,
			UpdatedAt: comment.UpdatedAt,
			ID:        comment.ID,
			Message:   comment.Message,
		})
	}
	if err := comments.Err(); err != nil {
		return fmt.Errorf("comments iteration: %w", err)
	}

	e := json.NewEncoder(rw)
	err = e.Encode(result)
	if err != nil {
		return fmt.Errorf("failed to encode results: %w", err)
	}

	return nil
}

func handlePostPost(rw http.ResponseWriter, req *http.Request, db *sqlx.DB) error {
	return errors.New("not implemented")
}
