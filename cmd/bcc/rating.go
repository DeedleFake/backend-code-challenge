package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/DeedleFake/backend-code-challenge/bcc"
	"github.com/jmoiron/sqlx"
)

func handlePostRating(req *http.Request, db *sqlx.DB) (interface{}, error) {
	var q struct {
		UserID  uint64  `json:"user_id"`
		RaterID uint64  `json:"rater_id"`
		Rating  float64 `json:"rating"`
	}
	err := json.NewDecoder(req.Body).Decode(&q)
	if err != nil {
		return nil, BadRequest(fmt.Errorf("parse body: %w", err))
	}

	if (q.Rating < 1) || (q.Rating > 5) {
		return nil, BadRequest(errors.New("rating must be between 1 and 5, inclusive"))
	}

	err = bcc.RateUser(db, q.RaterID, q.UserID, q.Rating)
	if err != nil {
		return nil, fmt.Errorf("rate user: %w", err)
	}

	return nil, nil
}
