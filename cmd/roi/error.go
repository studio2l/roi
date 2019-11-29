package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/studio2l/roi"
)

func handleError(w http.ResponseWriter, err error) {
	var e roi.Error
	if errors.As(err, &e) {
		if e.Log() != "" {
			log.Print(err)
		}
		http.Error(w, e.Error(), e.Code())
	}
	log.Print(err)
	http.Error(w, "unspecified internal error", http.StatusInternalServerError)
}
