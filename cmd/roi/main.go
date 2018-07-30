package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/securecookie"

	"golang.org/x/crypto/bcrypt"

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

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

func setSession(w http.ResponseWriter, userid string) error {
	value := map[string]string{
		"userid": userid,
	}
	encoded, err := cookieHandler.Encode("session", value)
	if err != nil {
		return err
	}
	c := &http.Cookie{
		Name:  "session",
		Value: encoded,
		Path:  "/",
	}
	http.SetCookie(w, c)
	return nil
}

func getSession(r *http.Request) (map[string]string, error) {
	c, _ := r.Cookie("session")
	if c == nil {
		return nil, nil
	}
	value := make(map[string]string)
	err := cookieHandler.Decode("session", c.Value, &value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func clearSession(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}
	recipt := struct {
		LoggedInUser string
	}{
		LoggedInUser: session["userid"],
	}
	err = executeTemplate(w, "index.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		id := r.Form.Get("id")
		if id == "" {
			http.Error(w, "id field emtpy", http.StatusBadRequest)
			return
		}
		pw := r.Form.Get("password")
		if pw == "" {
			http.Error(w, "password field emtpy", http.StatusBadRequest)
			return
		}
		db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error connecting to the database: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u, err := roi.GetUser(db, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not get user: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		if u.ID == "" {
			http.Error(w, fmt.Sprintf("user %s not exists", id), http.StatusInternalServerError)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(pw))
		if err != nil {
			http.Error(w, fmt.Sprintf("password not matched: %s", err), http.StatusInternalServerError)
			return
		}
		err = setSession(w, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not set session: %s", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}
	recipt := struct {
		LoggedInUser string
	}{
		LoggedInUser: session["userid"],
	}
	err = executeTemplate(w, "login.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		id := r.Form.Get("id")
		if id == "" {
			http.Error(w, "id field emtpy", http.StatusBadRequest)
			return
		}
		pw := r.Form.Get("password")
		if pw == "" {
			http.Error(w, "password field emtpy", http.StatusBadRequest)
			return
		}
		if len(pw) < 8 {
			http.Error(w, "password too short", http.StatusBadRequest)
			return
		}
		// 할일: password에 대한 컨펌은 프론트 엔드에서 하여야 함
		pwc := r.Form.Get("password_confirm")
		if pw != pwc {
			http.Error(w, "passwords are not matched", http.StatusBadRequest)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			// 할일: 어떤 상태를 전달하는 것이 가장 좋은가?
			http.Error(w, "password hashing error", http.StatusInternalServerError)
			return
		}
		db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
		if err != nil {
			fmt.Fprintln(os.Stderr, "error connecting to the database: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = roi.AddUser(db, id, string(hash))
		if err != nil {
			http.Error(w, fmt.Sprintf("could not add user: %s", err), http.StatusBadRequest)
			return
		}
		err = setSession(w, id)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not set session: %s", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}
	recipt := struct {
		LoggedInUser string
	}{
		LoggedInUser: session["userid"],
	}
	err = executeTemplate(w, "signup.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
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

	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}

	recipt := struct {
		LoggedInUser string
		Projects     []string
		Project      string
		Scenes       []string
		Shots        []roi.Shot
		FilterScene  string
		FilterShot   string
		FilterStatus string
	}{
		LoggedInUser: session["userid"],
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

func shotHandler(w http.ResponseWriter, r *http.Request) {
	pth := r.URL.Path[len("/shot/"):]
	pths := strings.Split(pth, "/")
	if len(pths) != 2 {
		http.NotFound(w, r)
		return
	}
	prj := pths[0]
	s := pths[1]
	db, err := sql.Open("postgres", "postgresql://maxroach@localhost:26257/roi?sslmode=disable")
	if err != nil {
		fmt.Fprintln(os.Stderr, "error connecting to the database: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shot, err := roi.FindShot(db, prj, s)
	if err != nil {
		log.Fatal(err)
	}
	if shot.Name == "" {
		http.NotFound(w, r)
		return
	}

	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}

	recipt := struct {
		LoggedInUser string
		Project      string
		Shot         roi.Shot
	}{
		LoggedInUser: session["userid"],
		Project:      prj,
		Shot:         shot,
	}
	err = executeTemplate(w, "shot.html", recipt)
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
	roi.CreateTableIfNotExists(db, "users", roi.UserTableFields)

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/login/", loginHandler)
	mux.HandleFunc("/logout/", logoutHandler)
	mux.HandleFunc("/signup/", signupHandler)
	mux.HandleFunc("/search/", searchHandler)
	mux.HandleFunc("/shot/", shotHandler)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServeTLS("0.0.0.0:443", "cert/cert.pem", "cert/key.pem", mux))
}
