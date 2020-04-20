package bcc

import (
	"time"

	"github.com/jmoiron/sqlx"
)

type GitHubEvent struct {
	ID        uint64    `db:"id" json:"id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UserID    uint64    `db:"user_id" json:"user_id"`

	Type     string `db:"type" json:"type"`
	RepoName string `db:"repo_name" json:"repo_name"`

	PRNumber *uint64 `db:"pr_number" json:"pr_number,omitempty"`

	NumCommits *int    `db:"num_commits" json:"num_commits,omitempty"`
	Head       *string `db:"head" json:"head,omitempty"`
}

func AddGitHubEvent(db *sqlx.DB, event GitHubEvent) error {
	_, err := db.Exec(
		`
		INSERT INTO github_events (
			id,
			created_at,
			user_id,
			type,
			repo_name,
			pr_number,
			num_commits,
			head
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT DO NOTHING
	`,
		event.ID,
		event.CreatedAt,
		event.UserID,
		event.Type,
		event.RepoName,
		event.PRNumber,
		event.NumCommits,
		event.Head,
	)
	return err
}
