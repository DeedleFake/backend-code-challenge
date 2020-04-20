// bcc-initdb is an admin tool for initializing the database.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

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

	db, err := sqlx.Open("postgres", fmt.Sprintf(
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
