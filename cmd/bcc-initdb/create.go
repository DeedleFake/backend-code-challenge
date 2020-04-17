package main

import "github.com/jmoiron/sqlx"

// createTables creates all of the required tables and sets all of the
// column options that are necessary. If reset is true, it drops any
// existing tables first.
func createTables(db *sqlx.DB, reset bool) (err error) {
	exec := func(stmt string, args ...interface{}) {
		if err != nil {
			return
		}

		_, err = db.Exec(stmt, args...)
	}

	if reset {
		exec(`DROP TABLE IF EXISTS users, posts, comments, ratings;`)
	}

	exec(`CREATE TABLE IF NOT EXISTS users (
		id serial NOT NULL PRIMARY KEY,
		registered_at timestamptz NOT NULL DEFAULT current_timestamp,
		created_at timestamptz NOT NULL DEFAULT current_timestamp,
		updated_at timestamptz NOT NULL DEFAULT current_timestamp,
		email text NOT NULL,
		name text NOT NULL,
		github_username text
	);`)
	exec(`CREATE TABLE IF NOT EXISTS posts (
		id serial NOT NULL PRIMARY KEY,
		user_id int NOT NULL,
		posted_at timestamptz NOT NULL DEFAULT current_timestamp,
		created_at timestamptz NOT NULL DEFAULT current_timestamp,
		updated_at timestamptz NOT NULL DEFAULT current_timestamp,
		title text NOT NULL,
		body text NOT NULL
	);`)
	exec(`CREATE TABLE IF NOT EXISTS comments (
		id serial NOT NULL PRIMARY KEY,
		user_id int NOT NULL,
		commented_at timestamptz NOT NULL DEFAULT current_timestamp,
		created_at timestamptz NOT NULL DEFAULT current_timestamp,
		updated_at timestamptz DEFAULT current_timestamp,
		post_id int NOT NULL,
		message text NOT NULL
	);`)
	exec(`CREATE TABLE IF NOT EXISTS ratings (
		id serial NOT NULL PRIMARY KEY,
		rated_at timestamptz NOT NULL DEFAULT current_timestamp,
		created_at timestamptz NOT NULL DEFAULT current_timestamp,
		updated_at timestamptz NOT NULL DEFAULT current_timestamp,
		user_id int NOT NULL,
		rater_id int NOT NULL,
		rating int NOT NULL
	);`)

	return err
}
