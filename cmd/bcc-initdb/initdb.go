package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func createTables(db *sql.DB, reset bool) (err error) {
	exec := func(stmt string, args ...interface{}) {
		if err != nil {
			return
		}

		_, err = db.Exec(stmt, args...)
	}

	if reset {
		exec(`drop table if exists users, posts, comments, ratings;`)
	}

	exec(`create table if not exists users (
		id int not null,
		registered_at timestamptz not null default current_timestamp,
		created_at timestamptz not null default current_timestamp,
		updated_at timestamptz not null default current_timestamp,
		email text not null,
		name text not null,
		github_username text,
		primary key (id)
	);`)
	exec(`create table if not exists posts (
		id int not null,
		user_id int not null,
		posted_at timestamptz not null default current_timestamp,
		created_at timestamptz not null default current_timestamp,
		updated_at timestamptz not null default current_timestamp,
		title text not null,
		body text not null,
		primary key (id)
	);`)
	exec(`create table if not exists comments (
		id int not null,
		user_id int not null,
		posted_at timestamptz not null default current_timestamp,
		created_at timestamptz not null default current_timestamp,
		updated_at timestamptz default current_timestamp,
		post_id int not null,
		message text not null,
		primary key (id)
	);`)
	exec(`create table if not exists ratings (
		id int not null,
		rated_at timestamptz not null default current_timestamp,
		created_at timestamptz not null default current_timestamp,
		updated_at timestamptz not null default current_timestamp,
		user_id int not null,
		rater_id int not null,
		rating int not null,
		primary key (id)
	);`)

	return err
}

func main() {
	addr := flag.String("dbaddr", "localhost", "Database address.")
	user := flag.String("dbuser", "postgres", "Database user.")
	pw := flag.String("dbpass", "", "Database password.")
	name := flag.String("dbname", "bcc", "Database name.")
	reset := flag.Bool("reset", false, "Reset all tables.")
	flag.Parse()

	db, err := sql.Open("postgres", fmt.Sprintf(
		"postgres://%v:%v@%v/%v?sslmode=disable",
		*user,
		*pw,
		*addr,
		*name,
	))
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	createTables(db, *reset)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}
