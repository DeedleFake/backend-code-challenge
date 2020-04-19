package main

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// createTables creates all of the required tables and sets all of the
// column options that are necessary. If reset is true, it drops any
// existing tables first.
func createTables(db *sqlx.DB, reset bool) (err error) {
	tables := []struct {
		name    string
		columns []string
	}{
		{
			name: "users",
			columns: []string{
				"id serial NOT NULL PRIMARY KEY",
				"registered_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"email text NOT NULL",
				"name text NOT NULL",
				"github_username text",
			},
		},
		{
			name: "posts",
			columns: []string{
				"id serial NOT NULL PRIMARY KEY",
				"user_id int NOT NULL",
				"posted_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"title text NOT NULL",
				"body text NOT NULL",
			},
		},
		{
			name: "comments",
			columns: []string{
				"id serial NOT NULL PRIMARY KEY",
				"user_id int NOT NULL",
				"commented_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"updated_at timestamptz DEFAULT CURRENT_TIMESTAMP",
				"post_id int NOT NULL",
				"message text NOT NULL",
			},
		},
		{
			name: "ratings",
			columns: []string{
				"id serial NOT NULL PRIMARY KEY",
				"rated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"user_id int NOT NULL",
				"rater_id int NOT NULL",
				"rating real NOT NULL",
			},
		},
		{
			name: "rating_events",
			columns: []string{
				"id serial NOT NULL PRIMARY KEY",
				"rated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"updated_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"rating_id int NOT NULL",
				"rating_before real NOT NULL",
				"rating_after real NOT NULL",
			},
		},
		{
			name: "github_events",
			columns: []string{
				"id text NOT NULL PRIMARY KEY",
				"created_at timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP",
				"user_id int NOT NULL",
				"type text NOT NULL",
				"repo_name text NOT NULL",
				"pr_number int",
				"num_commits int",
				"head text",
			},
		},
	}

	if reset {
		names := make([]string, 0, len(tables))
		for _, table := range tables {
			names = append(names, table.name)
		}
		_, err := db.Exec(`DROP TABLE IF EXISTS ` + strings.Join(names, ", "))
		if err != nil {
			return fmt.Errorf("drop tables: %w", err)
		}
	}

	for _, table := range tables {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS ` + table.name + ` (` + strings.Join(table.columns, ", ") + `)`)
		if err != nil {
			return fmt.Errorf("create %q: %w", table.name, err)
		}
	}

	return nil
}
