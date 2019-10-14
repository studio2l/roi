package main

import (
	"log"
	"net/http"

	"github.com/studio2l/roi"
)

func siteHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		log.Printf("could not get session: %s", err)
		clearSession(w)
	}
	ss, err := roi.Sites()
	if err != nil {
		log.Printf("could not get site: %s", err)
	}
	s := ss[0]
	recipt := struct {
		LoggedInUser string
		Site         *roi.Site
	}{
		LoggedInUser: session["userid"],
		Site:         s,
	}
	err = executeTemplate(w, "site.html", recipt)
	if err != nil {
		log.Fatal(err)
	}

}
