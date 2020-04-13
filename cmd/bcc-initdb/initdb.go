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
		registered timestamp not null,
		created timestamp not null,
		updated timestamp not null,
		email text not null,
		name text not null,
		github text,
		primary key (id)
	);`)
	exec(`create table if not exists posts (
		id int not null,
		user_id int not null,
		posted_at timestamp not null,
		created_at timestamp not null default current_timestamp,
		updated timestamp,
		title text not null,
		body text not null,
		primary key (id)
	);`)
	exec(`create table if not exists comments (
		id int not null,
		userid int not null,
		posted timestamp not null,
		created timestamp not null,
		updated timestamp,
		parentid int not null,
		title text not null,
		body text not null,
		primary key (id)
	);`)
	exec(`create table if not exists ratings (
		id int not null,
		rated timestamp not null,
		created timestamp not null,
		updated timestamp not null,
		userid int not null,
		raterid int not null,
		rating int not null,
		primary key (id)
	);`)

	return err
}

func main() {
	addr := flag.String("addr", "localhost", "Database address.")
	user := flag.String("user", "postgres", "Database user.")
	pw := flag.String("pass", "", "Database password.")
	name := flag.String("db", "bcc", "Database name.")
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
