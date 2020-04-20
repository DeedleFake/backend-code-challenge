// bcc runs a REST-ish API server.
//
// For a list of endpoints and their parameters, run bcc -doc.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sort"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func printDoc(endpoints map[APIMapping]APIEndpoint) {
	type sortable struct {
		M APIMapping
		H APIEndpoint
	}
	ep := make([]sortable, 0, len(endpoints))
	for m, h := range endpoints {
		ep = append(ep, sortable{M: m, H: h})
	}
	sort.Slice(ep, func(i1, i2 int) bool {
		if ep[i1].M.Path == ep[i2].M.Path {
			return ep[i1].M.Method < ep[i2].M.Method
		}
		return ep[i1].M.Path < ep[i2].M.Path
	})

	var sep string
	for _, endpoint := range ep {
		fmt.Printf("%v%v %v: %v\n", sep, endpoint.M.Method, endpoint.M.Path, endpoint.H.Desc())
		sep = "\n"

		params := reflect.Indirect(reflect.ValueOf(endpoint.H.Params()))
		paramsType := params.Type()
		for i := 0; i < params.NumField(); i++ {
			f := paramsType.Field(i)

			name := f.Name
			if tn := f.Tag.Get("query"); tn != "" {
				name = tn
			}
			if tn := f.Tag.Get("json"); tn != "" {
				name = tn
			}

			desc := ""
			if td := f.Tag.Get("desc"); td != "" {
				desc = ": " + td
			}

			fmt.Printf("\t%v (%v)%v\n", name, f.Type, desc)
		}
	}
}

func main() {
	doc := flag.Bool("doc", false, "show API documentation instead of starting server")
	addr := flag.String("addr", ":8080", "address to listen on")
	dbaddr := flag.String("dbaddr", "localhost", "database address")
	dbuser := flag.String("dbuser", "postgres", "database user")
	dbpass := flag.String("dbpass", "", "database password")
	dbname := flag.String("dbname", "bcc", "database name")
	flag.Parse()

	endpoints := map[APIMapping]APIEndpoint{
		{"GET", "/timeline"}: GetTimelineHandler{},

		{"GET", "/post"}:  GetPostHandler{},
		{"POST", "/post"}: PostPostHandler{},

		{"POST", "/comment"}:   PostCommentHandler{},
		{"DELETE", "/comment"}: DeleteCommentHandler{},

		{"POST", "/rating"}: PostRatingHandler{},
	}

	if *doc {
		printDoc(endpoints)
		return
	}

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

		Endpoints: endpoints,
	}

	log.Println("Starting server...")
	err = http.ListenAndServe(*addr, mux)
	log.Fatalf("Error starting server: %v", err)
}
