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
		if e.code == http.StatusInternalServerError {
			log.Print(e.msg)
			http.Error(w, "internal error", e.code)
			return
		}
		http.Error(w, e.msg, e.code)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
