package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"strings"
)

func handleMain(rw http.ResponseWriter, req *http.Request) {
	panic("Not implemented.")
}

func handleErrors(rw http.ResponseWriter, req *http.Request, err error) {
	var buf strings.Builder
	e := json.NewEncoder(&buf)
	e.Encode(map[string]interface{}{
		"type":  "error",
		"error": err.Error(),
	})

	http.Error(rw, buf.String(), http.StatusInternalServerError)
}

func main() {
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	http.Handle("/", ErrorHandler(LogHandler(http.HandlerFunc(handleMain)), handleErrors))

	log.Println("Starting server...")
	err := http.ListenAndServe(*addr, nil)
	log.Fatalf("Error starting server: %v", err)
}
