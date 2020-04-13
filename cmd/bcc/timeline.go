package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

func handleGetTimeline(rw http.ResponseWriter, req *http.Request, db *sql.DB) error {
	q := struct {
		UserID int `query:"user_id"`
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

	rows, err := db.Query(`
		SELECT posted_at, updated_at, title, body FROM posts
		WHERE user_id=$1
		ORDER BY posted_at
		LIMIT $2
	;`, q.UserID, q.Limit)
	if err != nil {
		return fmt.Errorf("select posts: %w", err)
	}
	defer rows.Close()

	type result struct {
		PostedAt  time.Time `json:"posted_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Title     string    `json:"title"`
		Body      string    `json:"body"`
	}

	results := []result{}
	for rows.Next() {
		var r result
		err = rows.Scan(&r.PostedAt, &r.UpdatedAt, &r.Title, &r.Body)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		results = append(results, r)
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
