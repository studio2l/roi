package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

var dev bool
var templates *template.Template

func parseTemplate() {
	templates = template.Must(template.ParseGlob("tmpl/*.html"))
}

func executeTemplate(w http.ResponseWriter, name string, data interface{}) {
	if dev {
		parseTemplate()
	}
	templates.ExecuteTemplate(w, name, data)
}
func rootHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, "index.html", nil)
}

func main() {
	dev = true

	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	if err := createTableIfNotExists(db, "rd7_shot", Shot{}); err != nil {
		log.Fatal(err)
	}
	s := Shot{
		Project: "test",
		Book:    5,
		Status:  "not good",
	}
	if err := insertInto(db, "rd7_shot", s); err != nil {
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	log.Fatal(http.ListenAndServe("0.0.0.0:7070", mux))
}
