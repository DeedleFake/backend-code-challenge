package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DeedleFake/backend-code-challenge/bcc"
	"github.com/jmoiron/sqlx"
)

type GetPostParams struct {
	PostID uint64 `query:"post_id"`
}

type GetPostHandler struct{}

func (h GetPostHandler) Desc() string {
	return "get a post and its comments"
}

func (h GetPostHandler) Params() interface{} {
	return &GetPostParams{}
}

func (h GetPostHandler) Serve(req *http.Request, db *sqlx.DB, params interface{}) (interface{}, error) {
	q := params.(*GetPostParams)

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
		UserID    uint64    `json:"user_id"`
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
			UserID    uint64    `json:"user_id"`
			PostedAt  time.Time `"json:"posted_at"`
			UpdatedAt time.Time `json:"updated_at"`
			ID        uint64    `json:"id"`
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

type PostPostParams struct {
	UserID *uint64 `json:"user_id"`
	Title  string  `json:"title"`
	Body   string  `json:"body"`
}

type PostPostHandler struct{}

func (h PostPostHandler) Desc() string {
	return "create a new post"
}

func (h PostPostHandler) Params() interface{} {
	return &PostPostParams{}
}

func (h PostPostHandler) Serve(req *http.Request, db *sqlx.DB, params interface{}) (interface{}, error) {
	q := params.(*PostPostParams)
	if q.UserID == nil {
		return nil, BadRequest(errors.New("user_id must be present"))
	}
	if q.Title == "" {
		return nil, BadRequest(errors.New("title must not be blank"))
	}

	err := bcc.CreatePost(db, *q.UserID, q.Title, q.Body)
	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	return nil, nil
}
