package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DeedleFake/backend-code-challenge/bcc"
	"github.com/jmoiron/sqlx"
)

type GetTimelineParams struct {
	UserID uint64 `query:"user_id"`
	Start  int    `query:"start" desc:"number of timeline entries to skip before returning results"`
	Limit  int    `query:"limit" desc:"maximum number of results to return"`
}

type GetTimelineHandler struct{}

func (h GetTimelineHandler) Desc() string {
	return "get a user's timeline"
}

func (h GetTimelineHandler) Params() interface{} {
	return &GetTimelineParams{
		Limit: 10,
	}
}

func (h GetTimelineHandler) Serve(req *http.Request, db *sqlx.DB, params interface{}) (interface{}, error) {
	q := params.(*GetTimelineParams)
	if q.Limit > 100 {
		return nil, BadRequest(errors.New("limit must not be larger than 100"))
	}

	entries, err := bcc.GetTimeline(db, q.UserID, q.Start, q.Limit)
	if err != nil {
		return nil, fmt.Errorf("get timeline: %w", err)
	}
	defer entries.Close()

	results := []bcc.TimelineEntry{}
	for entries.Next() {
		entry := entries.Current().(bcc.TimelineEntry)
		results = append(results, entry)
	}
	if err := entries.Err(); err != nil {
		return nil, fmt.Errorf("iteration: %w", err)
	}

	return results, nil
}
