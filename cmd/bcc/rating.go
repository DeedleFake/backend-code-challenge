package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DeedleFake/backend-code-challenge/bcc"
	"github.com/jmoiron/sqlx"
)

type PostRatingParams struct {
	UserID  uint64  `json:"user_id"`
	RaterID uint64  `json:"rater_id"`
	Rating  float64 `json:"rating"`
}

type PostRatingHandler struct{}

func (h PostRatingHandler) Desc() string {
	return "rate a user"
}

func (h PostRatingHandler) Params() interface{} {
	return &PostRatingParams{}
}

func (h PostRatingHandler) Serve(req *http.Request, db *sqlx.DB, params interface{}) (interface{}, error) {
	q := params.(*PostRatingParams)
	if (q.Rating < 1) || (q.Rating > 5) {
		return nil, BadRequest(errors.New("rating must be between 1 and 5, inclusive"))
	}

	err := bcc.RateUser(db, q.RaterID, q.UserID, q.Rating)
	if err != nil {
		return nil, fmt.Errorf("rate user: %w", err)
	}

	return nil, nil
}
