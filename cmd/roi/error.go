package main

import (
	"log"
	"net/http"
)

type httpError struct {
	msg  string
	code int
}

func (e httpError) Error() string {
	return e.msg
}

func handleError(w http.ResponseWriter, err error) {
	if e, ok := err.(httpError); ok {
		if e.code != http.StatusInternalServerError {
			http.Error(w, e.msg, e.code)
			return
		}
	}
	log.Print(err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}

func BadRequest(err error) httpError {
	return httpError{msg: err.Error(), code: http.StatusBadRequest}
}

func Internal(err error) httpError {
	return httpError{msg: err.Error(), code: http.StatusInternalServerError}
}
