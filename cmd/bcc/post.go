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

func handleGetPost(req *http.Request, db *sqlx.DB) (interface{}, error) {
	var q struct {
		PostID int `query:"post_id"`
	}
	err := parseQuery(req.URL.Query(), &q)
	if err != nil {
		return nil, APIUserError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to parse query: %w", err),
		}
	}

	post, err := bcc.GetPostByID(db, q.PostID)
	if err != nil {
		return nil, fmt.Errorf("post: %w", err)
	}

	comments, err := bcc.GetCommentsByPostID(db, q.PostID)
	if err != nil {
		return nil, fmt.Errorf("comments: %w", err)
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
		return nil, fmt.Errorf("comments iteration: %w", err)
	}

	return result, nil
}

func handlePostPost(req *http.Request, db *sqlx.DB) (interface{}, error) {
	var q struct {
		UserID *int   `json:"user_id"`
		Title  string `json:"title"`
		Body   string `json:"body"`
	}
	d := json.NewDecoder(req.Body)
	err := d.Decode(&q)
	if err != nil {
		return nil, APIUserError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to parse body: %w", err),
		}
	}
	if q.UserID == nil {
		return nil, APIUserError{
			Status: http.StatusBadRequest,
			Err:    errors.New("user_id must be present"),
		}
	}
	if q.Title == "" {
		return nil, APIUserError{
			Status: http.StatusBadRequest,
			Err:    errors.New("title must not be blank"),
		}
	}

	err = bcc.CreatePost(db, *q.UserID, q.Title, q.Body)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	return nil, nil
}
