package main

import (
	"fmt"
	"log"
	"net/http"
)

// ErrorHandlerFunc is used by ErrorHandler as an alternative handler
// when an error occurs.
type ErrorHandlerFunc func(rw http.ResponseWriter, req *http.Request, err error)

// ErrorHandler returns an http.Handler that handles panics that occur
// in h.
func ErrorHandler(h http.Handler, errh ErrorHandlerFunc) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		defer func() {
			perr := recover()
			if perr == nil {
				return
			}

			err, ok := perr.(error)
			if !ok {
				err = fmt.Errorf("%v", perr)
			}

			errh(rw, req, err)
		}()

		h.ServeHTTP(rw, req)
	})
}

// LogHandler returns an http.Handler that logs all requests that pass
// through it before handing them off to h.
func LogHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		log.Printf("%v %v", req.Method, req.URL.Path)
		h.ServeHTTP(rw, req)
	})
}
