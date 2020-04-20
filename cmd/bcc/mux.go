package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
)

// APIMapping combines an HTTP method and a URL path.
type APIMapping struct {
	Method, Path string
}

// APIEndpoint is an endpoint of the API.
type APIEndpoint interface {
	// Desc is a string describing the endpoint. It is purely for
	// documentation purposes.
	Desc() string

	// Params returns an instance of a type for holding the parameters
	// of this endpoint. For GET and DELETE requests, this will be
	// parsed into using parseQuery. For other request types, the body
	// of the request will be decoded into this object as JSON.
	Params() interface{}

	// Serve serves the endpoint to the client. The params are the value
	// returned by Params after having been filled. If err is nil then
	// rsp is encoded to JSON and returned to the client. If rsp and err
	// are nil, an empty object will be sent back.
	Serve(req *http.Request, db *sqlx.DB, params interface{}) (rsp interface{}, err error)
}

// APIMux implements a mux for API endpoints as an http.Handler.
type APIMux struct {
	// DB is the database connection to hand to the endpoint handlers.
	DB *sqlx.DB

	// Endpoints maps methods and paths to handlers.
	Endpoints map[APIMapping]APIEndpoint
}

func (mux APIMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Printf("%v %v", req.Method, req.URL.RequestURI())

	rw.Header().Set("Content-Type", "application/json")

	h := mux.Endpoints[APIMapping{
		Method: req.Method,
		Path:   req.URL.Path,
	}]
	if h == nil {
		http.Error(rw, `{"error": "invalid endpoint"}`, http.StatusNotFound)
		return
	}

	params := h.Params()
	switch req.Method {
	case "GET", "DELETE":
		err := parseQuery(req.URL.Query(), params)
		if err != nil {
			http.Error(rw, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
			return
		}

	default:
		err := json.NewDecoder(req.Body).Decode(params)
		if err != nil {
			http.Error(rw, fmt.Sprintf(`{"error":%q}`, err.Error()), http.StatusBadRequest)
		}
	}

	rsp, err := h.Serve(req, mux.DB, params)
	if err != nil {
		var errJSON string
		status := http.StatusInternalServerError

		var userErr APIUserError
		if errors.As(err, &userErr) {
			rsp, merr := json.Marshal(map[string]interface{}{"error": userErr.Error()})
			if merr != nil {
				log.Printf("Failed to marshal error: %v\nOriginal error: %v", merr, err)
				return
			}

			errJSON = string(rsp)
			if userErr.Status != 0 {
				status = userErr.Status
			}
		}

		http.Error(rw, errJSON, status)
		log.Printf("Error: %v", err)
		return
	}

	if rsp == nil {
		rsp = struct{}{}
	}

	err = json.NewEncoder(rw).Encode(rsp)
	if err != nil {
		log.Printf("Error sending response: %v", err)
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

// BadRequest returns a 400 error that wraps the given error.
func BadRequest(err error) error {
	return APIUserError{
		Status: http.StatusBadRequest,
		Err:    err,
	}
}
