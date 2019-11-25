package main

import "net/http"

type httpError struct {
	err  string
	code int
}

func (e httpError) Error() string {
	return e.err
}

func responseError(w http.ResponseWriter, err error) {
	if e, ok := err.(httpError); ok {
		http.Error(w, e.err, e.code)
		return
	}
	http.Error(w, err.Error(), http.StatusBadRequest)
}
