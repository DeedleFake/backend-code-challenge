package bcc

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// TimelineEntry is an entry in a user's timeline. Pointer fields may
// be null depending on the type of the entry. Valid types are "post",
// "comment", "passed_rating", and "github_event".
type TimelineEntry struct {
	Type string `db:"type" json:"type"`

	PostedAt  time.Time `db:"posted_at" json:"posted_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	ID        uint64    `db:"id" json:"id"`

	Title *string `db:"title" json:"title,omitempty"`
	Body  *string `db:"body" json:"body,omitempty"`

	PostID         *uint64  `db:"post_id" json:"post_id,omitempty"`
	Message        *string  `db:"message" json:"message,omitempty"`
	PostUserID     *uint64  `db:"post_user_id" json:"post_user_id,omitempty"`
	PostUserName   *string  `db:"post_user_name" json:"post_user_name,omitempty"`
	PostUserRating *float64 `db:"post_user_rating" json:"post_user_rating,omitempty"`

	PassedRatingBefore *float64 `db:"passed_rating_before" json:"passed_rating_before,omitempty"`
	PassedRatingAfter  *float64 `db:"passed_rating_after" json:"passed_rating_after,omitempty"`

	GitHubEventType    *string `db:"github_event_type" json:"github_event_type,omitempty"`
	GitHubEventRepo    *string `db:"github_event_repo" json:"github_event_repo,omitempty"`
	GitHubEventPR      *uint64 `db:"github_event_pr" json:"github_event_pr,omitempty"`
	GitHubEventCommits *int    `db:"github_event_commits" json:"github_event_commits,omitempty"`
	GitHubEventHead    *string `db:"github_event_head" json:"github_event_head,omitempty"`
}

// GetTimeline returns an iterator over the entries in a user's
// timeline, sorted in descending date order. start and limit control
// how many rows to return and where to start in the returned rows. In
// other words, a start of 10 and a limit of 20 will skip 10 rows and
// then return the 20 following those.
func GetTimeline(db *sqlx.DB, userID uint64, start, limit int) (*Iterator, error) {
	rows, err := db.Queryx(`
		SELECT
			'post' AS type,
			posted_at,
			updated_at,
			id,
			title,
			body,
			NULL AS message,
			NULL AS post_id,
			NULL AS post_user_id,
			NULL AS post_user_name,
			NULL AS post_user_rating,
			NULL :: real AS passed_rating_before,
			NULL :: real AS passed_rating_after,
			NULL :: text AS github_event_type,
			NULL :: text AS github_event_repo,
			NULL :: bigint AS github_event_pr,
			NULL :: bigint AS github_event_commits,
			NULL :: text AS github_event_head
		FROM posts
			WHERE user_id = $1

		UNION ALL

		SELECT
			'comment' AS type,
			commented_at AS posted_at,
			comments.updated_at AS updated_at,
			comments.id AS id,
			NULL AS title,
			NULL AS body,
			message,
			post_id,
			users.id AS post_user_id,
			users.name AS post_user_name,
			(
				SELECT AVG(rating) FROM (
					SELECT
						ROW_NUMBER() OVER (PARTITION BY rater_id ORDER BY rated_at DESC) AS rn,
						rating
					FROM ratings
						WHERE user_id = posts.user_id
				) AS r WHERE rn=1
			) AS post_user_rating,
			NULL AS passed_rating_before,
			NULL AS passed_rating_after,
			NULL AS github_event_type,
			NULL AS github_event_repo,
			NULL AS github_event_pr,
			NULL AS github_event_commits,
			NULL AS github_event_head
		FROM comments
			JOIN posts ON posts.id = comments.post_id
			JOIN users ON users.id = posts.user_id
			WHERE comments.user_id = $1

		UNION ALL

		SELECT
			'passed_rating' AS type,
			rating_events.rated_at AS posted_at,
			rating_events.updated_at AS updated_at,
			rating_events.id AS id,
			NULL AS title,
			NULL AS body,
			NULL AS message,
			NULL AS post_id,
			NULL AS post_user_id,
			NULL AS post_user_name,
			NULL AS post_user_rating,
			rating_events.rating_before AS passed_rating_before,
			rating_events.rating_after AS passed_rating_after,
			NULL AS github_event_type,
			NULL AS github_event_repo,
			NULL AS github_event_pr,
			NULL AS github_event_commits,
			NULL AS github_event_head
		FROM rating_events
			JOIN ratings ON rating_events.rating_id = ratings.id
			WHERE user_id = $1
				AND rating_events.rating_before < 4
				AND rating_events.rating_after >= 4

		UNION ALL

		SELECT
			'github_event' AS type,
			github_events.created_at AS posted_at,
			github_events.created_at AS updated_at,
			github_events.id AS id,
			NULL AS title,
			NULL AS body,
			NULL AS message,
			NULL AS post_id,
			NULL AS post_user_id,
			NULL AS post_user_name,
			NULL AS post_user_rating,
			NULL AS passed_rating_before,
			NULL AS passed_rating_after,
			github_events.type AS github_event_type,
			github_events.repo_name AS github_event_repo,
			github_events.pr_number AS github_event_pr,
			github_events.num_commits AS github_event_commits,
			github_events.head AS github_event_head
		FROM github_events
			WHERE user_id = $1

		ORDER BY posted_at DESC
		LIMIT $3 OFFSET $2
	`, userID, start, limit)
	if err != nil {
		return nil, err
	}

	return &Iterator{
		next: rows.Next,
		cur: func() (interface{}, error) {
			var entry TimelineEntry
			err := rows.StructScan(&entry)
			return entry, err
		},
		close: rows.Close,
	}, nil
}
