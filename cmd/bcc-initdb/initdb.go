package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

// createTables creates all of the required tables and sets all of the
// column options that are necessary. If reset is true, it drops any
// existing tables first.
func createTables(db *sql.DB, reset bool) (err error) {
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

// insertData reads the CSV file at path and inserts each row of the
// file into the given table. It expects the first row of the CSV to
// be a list of the columns that correspond to the columns of the CSV
// file. In other words, the file
//
//    col1,col2
//    an,example
//    and,another
//
// will attempt to insert "an" and "and" into the column "col1" and
// "example" and "another" into the column "col2".
//
// As this entire program is strictly an admin tool, any data being
// inserted by this tool is assumed to be trusted. In particular, the
// column names from the CSV file are manually inserted into the SQL.
func insertData(db *sql.DB, table, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.LazyQuotes = true
	cols, err := r.Read()
	if err != nil {
		return fmt.Errorf("read columns: %w", err)
	}

	args := make([]string, 0, len(cols))
	for i := range cols {
		args = append(args, fmt.Sprintf("$%v", i+1))
	}

	insert, err := db.Prepare(`INSERT INTO ` + table + ` (` + strings.Join(cols, ", ") + `) VALUES (` + strings.Join(args, ", ") + `);`)
	if err != nil {
		return fmt.Errorf("prepare: %w", err)
	}
	defer insert.Close()

	for {
		row, err := r.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return fmt.Errorf("read row: %w", err)
		}

		args := make([]interface{}, 0, len(row))
		for _, c := range row {
			args = append(args, c)
		}

		_, err = insert.Exec(args...)
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
	}

	return nil
}

// dataFlag is an implementation of flag.Value that reads a
// comma-separated list of key=value pairs.
type dataFlag map[string]string

func (df dataFlag) String() string {
	var sb strings.Builder

	var sep string
	for k, v := range df {
		fmt.Fprintf(&sb, "%v%v=%v", sep, k, v)
		sep = ","
	}

	return sb.String()
}

func (df *dataFlag) Set(val string) error {
	if *df == nil {
		*df = make(dataFlag)
	}

	pairs := strings.Split(val, ",")
	for _, pair := range pairs {
		split := strings.SplitN(pair, "=", 2)
		if len(split) < 2 {
			return fmt.Errorf("%q is not valid", pair)
		}

		(*df)[split[0]] = split[1]
	}

	return nil
}

func main() {
	addr := flag.String("dbaddr", "localhost", "Database address")
	user := flag.String("dbuser", "postgres", "Database user")
	pw := flag.String("dbpass", "", "Database password")
	name := flag.String("dbname", "bcc", "Database name")
	reset := flag.Bool("reset", false, "Reset all tables")

	var data dataFlag
	flag.Var(&data, "data", "Comma separated list of table names and CSV files with data to insert into them")

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

	var wg sync.WaitGroup
	for table, path := range data {
		wg.Add(1)
		go func(table, path string) {
			defer wg.Done()

			err := insertData(db, table, path)
			if err != nil {
				log.Printf("Failed to insert data into %q: %v", table, err)
			}
		}(table, path)
	}
	wg.Wait()
}
