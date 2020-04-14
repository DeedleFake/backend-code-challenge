package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

func handleGetTimeline(rw http.ResponseWriter, req *http.Request, db *sqlx.DB) error {
	q := struct {
		UserID int `query:"user_id"`
		Start  int `query:"start"`
		Limit  int `query:"limit"`
	}{
		Limit: 10,
	}
	err := parseQuery(req.URL.Query(), &q)
	if err != nil {
		return APIUserError{
			Status: http.StatusBadRequest,
			Err:    fmt.Errorf("failed to parse query: %w", err),
		}
	}

	if q.Limit > 50 {
		return APIUserError{
			Status: http.StatusBadRequest,
			Err:    errors.New("limit must not be larger than 50"),
		}
	}

	rows, err := db.Queryx(`
		SELECT
			'post' AS type,
			posted_at,
			updated_at,
			id,
			title,
			body,
			NULL AS post_id,
			NULL AS message
		FROM posts
		UNION ALL
		SELECT
			'comment' AS type,
			commented_at AS posted_at,
			updated_at,
			id,
			NULL AS title,
			NULL AS body,
			post_id,
			message
		FROM comments
		WHERE user_id=$1
		ORDER BY posted_at DESC
		LIMIT $2 OFFSET $3
	;`, q.UserID, q.Limit, q.Start)
	if err != nil {
		return fmt.Errorf("select posts: %w", err)
	}
	defer rows.Close()

	type resultPost struct {
		ID        int       `json:"id"`
		PostedAt  time.Time `json:"posted_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
	}

	type resultComment struct {
		ID          int       `json:"id"`
		CommentedAt time.Time `json:"commented_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		PostID      int       `json:"post_id"`
		Message     string    `json:"message"`
	}

	results := []interface{}{}
	for rows.Next() {
		var result struct {
			Type string `db:"type"`

			PostedAt  time.Time `db:"posted_at"`
			UpdatedAt time.Time `db:"updated_at"`
			ID        int       `db:"id"`

			Title *string `db:"title"`
			Body  *string `db:"body"`

			PostID  *int    `db:"post_id"`
			Message *string `db:"message"`
		}
		err = rows.StructScan(&result)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		switch result.Type {
		case "post":
			results = append(results, resultPost{
				ID:        result.ID,
				PostedAt:  result.PostedAt,
				UpdatedAt: result.UpdatedAt,
				Title:     *result.Title,
				Body:      *result.Body,
			})

		case "comment":
			results = append(results, resultComment{
				ID:          result.ID,
				CommentedAt: result.PostedAt,
				UpdatedAt:   result.UpdatedAt,
				PostID:      *result.PostID,
				Message:     *result.Message,
			})

		default:
			panic(fmt.Errorf("unexpected result type: %q", result.Type))
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to advance rows: %w", err)
	}

	e := json.NewEncoder(rw)
	err = e.Encode(results)
	if err != nil {
		return fmt.Errorf("failed to encode results: %w", err)
	}

	return nil
}
