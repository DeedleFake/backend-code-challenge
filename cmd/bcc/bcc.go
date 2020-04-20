package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	dbaddr := flag.String("dbaddr", "localhost", "Database address")
	dbuser := flag.String("dbuser", "postgres", "Database user")
	dbpass := flag.String("dbpass", "", "Database password")
	dbname := flag.String("dbname", "bcc", "Database name")
	flag.Parse()

	db, err := sqlx.Open("postgres", fmt.Sprintf(
		"postgres://%v:%v@%v/%v?sslmode=disable",
		*dbuser,
		*dbpass,
		*dbaddr,
		*dbname,
	))
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	mux := &APIMux{
		DB: db,

		Endpoints: map[APIMapping]APIEndpoint{
			{"GET", "/timeline"}: GetTimelineHandler{},

			{"GET", "/post"}:  GetPostHandler{},
			{"POST", "/post"}: PostPostHandler{},

			{"POST", "/rating"}: PostRatingHandler{},
		},
	}

	log.Println("Starting server...")
	err = http.ListenAndServe(*addr, mux)
	log.Fatalf("Error starting server: %v", err)
}
