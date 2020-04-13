package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "github.com/lib/pq"
)

// APIMapping combines an HTTP method and a URL path.
type APIMapping struct {
	Method, Path string
}

// APIEndpoint is the function signature of APIMux handlers.
type APIEndpoint func(rw http.ResponseWriter, req *http.Request, db *sql.DB) error

// APIMux implements a mux for API endpoints as an http.Handler.
type APIMux struct {
	// DB is the database connection to hand to the endpoint handlers.
	DB *sql.DB

	// Endpoints maps methods and paths to handlers. Methods are
	// converted to lowercase before being checked against this map, so
	// any entries added here should use lowercase for specifying their
	// mappings. Paths are not modified.
	Endpoints map[APIMapping]APIEndpoint
}

func (mux APIMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Printf("%v %v", req.Method, req.URL.RequestURI())

	rw.Header().Set("Content-Type", "application/json")

	req.Method = strings.ToLower(req.Method)

	h := mux.Endpoints[APIMapping{
		Method: req.Method,
		Path:   req.URL.Path,
	}]
	if h == nil {
		http.Error(rw, `{"error": "invalid endpoint"}`, http.StatusNotFound)
		return
	}

	err := h(rw, req, mux.DB)
	if err != nil {
		var userErr APIUserError
		if errors.As(err, &userErr) {
			rsp, merr := json.Marshal(map[string]interface{}{"error": userErr.Error()})
			if merr != nil {
				log.Printf("Failed to marshal error: %v\nOriginal error: %v", merr, err)
				return
			}

			status := userErr.Status
			if status == 0 {
				status = http.StatusInternalServerError
			}
			http.Error(rw, string(rsp), status)
		}

		log.Printf("Error: %v", err)
	}
}

// APIUserError is returned by APIEndpoints that want to send error
// data back to the user. If Status is zero, it is presumed to be
// StatusInternalServerError.
type APIUserError struct {
	Status int
	Err    error
}

func (err APIUserError) Error() string {
	return err.Err.Error()
}

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	dbaddr := flag.String("dbaddr", "localhost", "Database address")
	dbuser := flag.String("dbuser", "postgres", "Database user")
	dbpass := flag.String("dbpass", "", "Database password")
	dbname := flag.String("dbname", "bcc", "Database name")
	flag.Parse()

	db, err := sql.Open("postgres", fmt.Sprintf(
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
			{"get", "/timeline"}: handleGetTimeline,
		},
	}

	log.Println("Starting server...")
	err = http.ListenAndServe(*addr, mux)
	log.Fatalf("Error starting server: %v", err)
}
