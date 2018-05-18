package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"sort"

	"dev.2lfilm.com/2l/roi"
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
	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	prj := "test"

	shots, err := roi.SelectShots(db, prj, nil)
	if err != nil {
		log.Fatal(err)
	}
	sort.Slice(shots, func(i int, j int) bool {
		if shots[i].Project < shots[j].Project {
			return true
		}
		if shots[i].Project > shots[j].Project {
			return false
		}
		if shots[i].Scene < shots[j].Scene {
			return true
		}
		if shots[i].Scene > shots[j].Scene {
			return false
		}
		return shots[i].Name <= shots[j].Name
	})

	recipt := struct {
		Project string
		Shots   []roi.Shot
	}{
		Project: prj,
		Shots:   shots,
	}
	executeTemplate(w, "index.html", recipt)
}

func shotHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Path[len("/shot/"):]
	if code == "" {
		return
	}

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
	if len(prjs) == 0 {
		fmt.Fprintln(os.Stderr, "no projects")
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
	for _, k := range []string{"book", "scene", "status"} {
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
		FilterBook   string
		FilterScene  string
		FilterStatus string
	}{
		Projects:     prjs,
		Project:      code,
		Scenes:       scenes,
		Shots:        shots,
		FilterBook:   where["book"],
		FilterScene:  where["scene"],
		FilterStatus: where["status"],
	}
	executeTemplate(w, "index.html", recipt)
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
	mux.HandleFunc("/shot/", shotHandler)
	log.Fatal(http.ListenAndServe("0.0.0.0:7070", mux))
}
