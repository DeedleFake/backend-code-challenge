package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

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
func insertData(db *sqlx.DB, table, path string) error {
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
			a := interface{}(c)
			if a == "" {
				a = nil
			}
			args = append(args, a)
		}

		_, err = insert.Exec(args...)
		if err != nil {
			return fmt.Errorf("insert: %w", err)
		}
	}

	types, err := db.QueryRowx(`SELECT * FROM ` + table).ColumnTypes()
	if err != nil {
		return fmt.Errorf("column types: %w", err)
	}

	for _, col := range types {
		switch col.DatabaseTypeName() {
		case "INT4":
			var max int
			err := db.QueryRowx(`SELECT MAX(` + col.Name() + `) FROM ` + table).Scan(&max)
			if err != nil {
				return fmt.Errorf("max %q: %w", col.Name(), err)
			}
			err = db.QueryRowx(`SELECT setval(pg_get_serial_sequence($1, $2), $3)`, table, col.Name(), max).Err()
			if err != nil {
				return fmt.Errorf("setval %q: %w", col.Name(), err)
			}
		}
	}

	return nil
}
