package bcc

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type TimelineEntry struct {
	Type string `db:"type" json:"type"`

	PostedAt  time.Time `db:"posted_at" json:"posted_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	ID        int       `db:"id" json:"id"`

	Title *string `db:"title" json:"title,omitempty"`
	Body  *string `db:"body" json:"body,omitempty"`

	PostID         *int     `db:"post_id" json:"post_id,omitempty"`
	Message        *string  `db:"message" json:"message,omitempty"`
	PostUserID     *int     `db:"post_user_id" json:"post_user_id"`
	PostUserName   *string  `db:"post_user_name" json:"post_user_name"`
	PostUserRating *float64 `db:"post_user_rating" json:"post_user_rating"`
}

func GetTimeline(db *sqlx.DB, userID, start, limit int) (*Iterator, error) {
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
			NULL AS post_user_rating
		FROM posts
			WHERE user_id=$1
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
			) AS post_user_rating
		FROM comments
			JOIN posts ON (posts.id = comments.post_id)
			JOIN users ON (users.id = posts.user_id)
			WHERE comments.user_id=$1
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
