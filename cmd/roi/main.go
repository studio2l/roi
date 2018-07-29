package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"dev.2lfilm.com/2l/roi"
)

var dev bool
var templates *template.Template

func parseTemplate() {
	templates = template.Must(template.ParseGlob("tmpl/*.html"))
}

func executeTemplate(w http.ResponseWriter, name string, data interface{}) error {
	if dev {
		parseTemplate()
	}
	return templates.ExecuteTemplate(w, name, data)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	executeTemplate(w, "index.html", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/search/"):]

	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error connecting to the database: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	prjRows, err := db.Query("SELECT code FROM projects")
	if err != nil {
		fmt.Fprintln(os.Stderr, "project selection error: ", err)
		return
	}
	defer prjRows.Close()
	prjs := make([]string, 0)
	for prjRows.Next() {
		prj := ""
		if err := prjRows.Scan(&prj); err != nil {
			fmt.Fprintln(os.Stderr, "error getting prject info from database: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		prjs = append(prjs, prj)
	}

	if code == "" && len(prjs) != 0 {
		// 할일: 추후 사용자가 마지막으로 선택했던 프로젝트로 이동
		http.Redirect(w, r, "/search/"+prjs[0], http.StatusSeeOther)
		return
	}
	found := false
	for _, p := range prjs {
		if p == code {
			found = true
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "not found project %s\n", code)
		return
		// http.Error(w, fmt.Sprintf("not found project: %s", code), http.StatusNotFound)
	}

	scenes, err := roi.SelectScenes(db, code)
	if err != nil {
		log.Fatal(err)
	}

	where := make(map[string]string)
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	for _, k := range []string{"scene", "shot", "status"} {
		v := r.Form.Get(k)
		if v != "" {
			where[k] = v
		}
	}
	fmt.Println(where)
	shots, err := roi.SelectShots(db, code, where)
	if err != nil {
		log.Fatal(err)
	}

	recipt := struct {
		Projects     []string
		Project      string
		Scenes       []string
		Shots        []roi.Shot
		FilterScene  string
		FilterShot   string
		FilterStatus string
	}{
		Projects:     prjs,
		Project:      code,
		Scenes:       scenes,
		Shots:        shots,
		FilterScene:  where["scene"],
		FilterShot:   where["shot"],
		FilterStatus: where["status"],
	}
	err = executeTemplate(w, "search.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	dev = true

	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	roi.CreateTableIfNotExists(db, "projects", roi.ProjectTableFields)

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/search/", searchHandler)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe("0.0.0.0:7070", mux))
}
