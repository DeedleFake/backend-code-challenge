package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/DeedleFake/backend-code-challenge/bcc"
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

	if q.Limit > 100 {
		return APIUserError{
			Status: http.StatusBadRequest,
			Err:    errors.New("limit must not be larger than 100"),
		}
	}

	entries, err := bcc.GetTimeline(db, q.UserID, q.Start, q.Limit)
	if err != nil {
		return fmt.Errorf("get timeline: %w", err)
	}
	defer entries.Close()

	var results []bcc.TimelineEntry
	for entries.Next() {
		entry := entries.Current().(bcc.TimelineEntry)
		results = append(results, entry)
	}
	if err := entries.Err(); err != nil {
		return fmt.Errorf("iteration: %w", err)
	}

	e := json.NewEncoder(rw)
	err = e.Encode(results)
	if err != nil {
		return fmt.Errorf("failed to encode results: %w", err)
	}

	return nil
}
