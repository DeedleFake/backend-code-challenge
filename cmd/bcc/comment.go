package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DeedleFake/backend-code-challenge/bcc"
	"github.com/jmoiron/sqlx"
)

type PostCommentParams struct {
	UserID  uint64 `json:"user_id" desc:"ID of the user making the comment"`
	PostID  uint64 `json:"post_id" desc:"ID of the post on which a comment is being made"`
	Message string `json:"message" desc:"contents of the comment"`
}

type PostCommentHandler struct{}

func (h PostCommentHandler) Desc() string {
	return "make a comment on a post"
}

func (h PostCommentHandler) Params() interface{} {
	return &PostCommentParams{}
}

func (h PostCommentHandler) Serve(req *http.Request, db *sqlx.DB, params interface{}) (interface{}, error) {
	q := params.(*PostCommentParams)
	if q.Message == "" {
		return nil, BadRequest(errors.New("message must not be blank"))
	}

	err := bcc.CreateComment(db, q.UserID, q.PostID, q.Message)
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	return nil, nil
}

type DeleteCommentParams struct {
	CommentID uint64 `query:"comment_id" desc:"ID of the comment being deleted"`
}

type DeleteCommentHandler struct{}

func (h DeleteCommentHandler) Desc() string {
	return "delete a comment"
}

func (h DeleteCommentHandler) Params() interface{} {
	return &DeleteCommentParams{}
}

func (h DeleteCommentHandler) Serve(req *http.Request, db *sqlx.DB, params interface{}) (interface{}, error) {
	q := params.(*DeleteCommentParams)

	err := bcc.DeleteComment(db, q.CommentID)
	if err != nil {
		return nil, fmt.Errorf("delete comment: %w", err)
	}

	return nil, nil
}
