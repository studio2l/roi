package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/securecookie"

	"github.com/studio2l/roi"
)

// dev는 현재 개발모드인지를 나타낸다.
var dev bool

// templates에는 사용자에게 보일 페이지의 템플릿이 담긴다.
var templates *template.Template

// hasThumbnail은 해당 특정 프로젝트 샷에 썸네일이 있는지 검사한다.
//
// 주의: 만일 썸네일 파일 검사시 에러가 나면 이 함수는 썸네일이 있다고 판단한다.
// 이 함수는 템플릿 안에서 쓰이기 때문에 프론트 엔드에서 한번 더 검사하게
// 만들기 위해서이다.
func hasThumbnail(prj, shot string) bool {
	_, err := os.Stat(fmt.Sprintf("roi-userdata/thumbnail/%s/%s.png", prj, shot))
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return true // 함수 주석 참고
	}
	return true
}

// stringFromTime은 시간을 rfc3339 형식의 문자열로 표현한다.
func stringFromTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format(time.RFC3339)
}

// stringFromDate는 시간을 rfc3339 형식의 문자열로 표현하되 T부터는 표시하지 않는다.
func stringFromDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return strings.Split(t.Local().Format(time.RFC3339), "T")[0]
}

// timeFromString는 rfc3339 형식의 문자열에서 시간을 얻는다.
// 받은 문자열이 형식에 맞지 않으면 에러를 반환한다.
func timeFromString(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// parseTimeforms는 http.Request.Form에서 시간 형식의 Form에 대해 파싱해
// 맵으로 반환한다. 만일 받아들인 문자열이 시간 형식에 맞지 않으면 에러를 낸다.
func parseTimeForms(form url.Values, keys ...string) (map[string]time.Time, error) {
	tforms := make(map[string]time.Time)
	for _, k := range keys {
		v := form.Get(k)
		if v == "" {
			continue
		}
		t, err := timeFromString(v)
		if err != nil {
			return nil, fmt.Errorf("invalid time string '%s' for '%s'", v, k)
		}
		tforms[k] = t
	}
	return tforms, nil
}

// parseTemplate은 tmpl 디렉토리 안의 html파일들을 파싱하여 http 응답에 사용될 수 있도록 한다.
func parseTemplate() {
	templates = template.Must(template.New("").Funcs(template.FuncMap{
		"hasThumbnail":   hasThumbnail,
		"stringFromTime": stringFromTime,
		"stringFromDate": stringFromDate,
		"join":           strings.Join,
	}).ParseGlob("tmpl/*.html"))
}

// executeTemplate은 템플릿과 정보를 이용하여 w에 응답한다.
// templates.ExecuteTemplate 대신 이 함수를 쓰는 이유는 개발모드일 때
// 재 컴파일 없이 업데이트된 템플릿을 사용할 수 있기 때문이다.
func executeTemplate(w http.ResponseWriter, name string, data interface{}) error {
	if dev {
		parseTemplate()
	}
	return templates.ExecuteTemplate(w, name, data)
}

// cookieHandler는 클라이언트 브라우저 세션에 암호화된 쿠키를 저장을 돕는다.
var cookieHandler *securecookie.SecureCookie

// setSession은 클라이언트 브라우저에 세션을 저장한다.
func setSession(w http.ResponseWriter, session map[string]string) error {
	encoded, err := cookieHandler.Encode("session", session)
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

// getSession은 클라이언트 브라우저에 저장되어 있던 세션을 불러온다.
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

// clearSession은 클라이언트 브라우저에 저장되어 있던 세션을 지운다.
func clearSession(w http.ResponseWriter) {
	c := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, c)
}

// rootHandler는 루트경로(/)를 포함해 정의되지 않은 페이지로의 사용자 접속을 처리한다.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}
	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}
	if session == nil || session["userid"] == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	user := session["userid"]
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	tasks, err := roi.UserTasks(db, user)
	if err != nil {
		log.Printf("could not get user tasks: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	numTasks := make(map[string]map[roi.TaskStatus]int)
	for _, t := range tasks {
		if numTasks[t.ProjectID] == nil {
			numTasks[t.ProjectID] = make(map[roi.TaskStatus]int)
		}
		numTasks[t.ProjectID][t.Status] += 1
	}
	recipt := struct {
		LoggedInUser string
		NumTasks     map[string]map[roi.TaskStatus]int
	}{
		LoggedInUser: session["userid"],
		NumTasks:     numTasks,
	}
	err = executeTemplate(w, "index.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

// loginHandler는 /login 페이지로 사용자가 접속했을때 로그인 페이지를 반환한다.
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
		db, err := roi.DB()
		if err != nil {
			log.Printf("could not connect to database: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		match, err := roi.UserPasswordMatch(db, id, pw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !match {
			http.Error(w, "entered password is not correct", http.StatusBadRequest)
			return
		}
		session := map[string]string{
			"userid": id,
		}
		err = setSession(w, session)
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

// logoutHandler는 /logout 페이지로 사용자가 접속했을때 사용자를 로그아웃 시킨다.
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// signupHandler는 /signup 페이지로 사용자가 접속했을때 가입 페이지를 반환한다.
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
		db, err := roi.DB()
		if err != nil {
			log.Printf("could not connect to database: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		err = roi.AddUser(db, id, pw)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not add user: %s", err), http.StatusBadRequest)
			return
		}
		session := map[string]string{
			"userid": id,
		}
		err = setSession(w, session)
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

// profileHandler는 /profile 페이지로 사용자가 접속했을 때 사용자 프로필 페이지를 반환한다.
func profileHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
		http.Redirect(w, r, "/login/", http.StatusSeeOther)
		return
	}
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if r.Method == "POST" {
		r.ParseForm()
		upd := roi.UpdateUserParam{
			KorName:     r.Form.Get("kor_name"),
			Name:        r.Form.Get("name"),
			Team:        r.Form.Get("team"),
			Role:        r.Form.Get("position"),
			Email:       r.Form.Get("email"),
			PhoneNumber: r.Form.Get("phone_number"),
			EntryDate:   r.Form.Get("entry_date"),
		}
		err = roi.UpdateUser(db, session["userid"], upd)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not set user: %s", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings/profile", http.StatusSeeOther)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get user: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	fmt.Println(u)
	recipt := struct {
		LoggedInUser string
		User         *roi.User
	}{
		LoggedInUser: session["userid"],
		User:         u,
	}
	err = executeTemplate(w, "profile.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

// updatePasswordHandler는 /update-password 페이지로 사용자가 패스워드 변경과 관련된 정보를 보내면
// 사용자 패스워드를 변경한다.
func updatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get session: %s", err), http.StatusInternalServerError)
		clearSession(w)
		return
	}
	r.ParseForm()
	oldpw := r.Form.Get("old_password")
	if oldpw == "" {
		http.Error(w, "old password field emtpy", http.StatusBadRequest)
		return
	}
	newpw := r.Form.Get("new_password")
	if newpw == "" {
		http.Error(w, "new password field emtpy", http.StatusBadRequest)
		return
	}
	if len(newpw) < 8 {
		http.Error(w, "new password too short", http.StatusBadRequest)
		return
	}
	// 할일: password에 대한 컨펌은 프론트 엔드에서 하여야 함
	newpwc := r.Form.Get("new_password_confirm")
	if newpw != newpwc {
		http.Error(w, "passwords are not matched", http.StatusBadRequest)
		return
	}
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	id := session["userid"]
	match, err := roi.UserPasswordMatch(db, id, oldpw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !match {
		http.Error(w, "entered password is not correct", http.StatusBadRequest)
		return
	}
	err = roi.UpdateUserPassword(db, id, newpw)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not change user password: %s", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusSeeOther)
}

// projectsHandler는 /project 페이지로 사용자가 접속했을때 페이지를 반환한다.
func projectsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	prjs, err := roi.AllProjects(db)
	if err != nil {
		log.Print(fmt.Sprintf("error while getting projects: %s", err))
		return
	}

	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}

	recipt := struct {
		LoggedInUser string
		Projects     []*roi.Project
	}{
		LoggedInUser: session["userid"],
		Projects:     prjs,
	}
	err = executeTemplate(w, "projects.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

// addProjectHandler는 /add-project 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트를 생성한다.
func addProjectHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if u.Role != "admin" {
		// 할일: admin이 아닌 사람은 프로젝트를 생성할 수 없도록 하기
	}
	if r.Method == "POST" {
		r.ParseForm()
		id := r.Form.Get("id")
		if id == "" {
			http.Error(w, "need project 'id'", http.StatusBadRequest)
			return
		}
		exist, err := roi.ProjectExist(db, id)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if exist {
			http.Error(w, fmt.Sprintf("project '%s' exist", id), http.StatusBadRequest)
			return
		}
		timeForms, err := parseTimeForms(r.Form,
			"start_date",
			"release_date",
			"crank_in",
			"crank_up",
			"vfx_due_date",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		p := &roi.Project{
			ID:            id,
			Name:          r.Form.Get("name"),
			Status:        "waiting",
			Client:        r.Form.Get("client"),
			Director:      r.Form.Get("director"),
			Producer:      r.Form.Get("producer"),
			VFXSupervisor: r.Form.Get("vfx_supervisor"),
			VFXManager:    r.Form.Get("vfx_manager"),
			CGSupervisor:  r.Form.Get("cg_supervisor"),
			StartDate:     timeForms["start_date"],
			ReleaseDate:   timeForms["release_date"],
			CrankIn:       timeForms["crank_in"],
			CrankUp:       timeForms["crank_up"],
			VFXDueDate:    timeForms["vfx_due_date"],
			OutputSize:    r.Form.Get("output_size"),
			ViewLUT:       r.Form.Get("view_lut"),
			DefaultTasks:  fields(r.Form.Get("default_tasks"), ","),
		}
		err = roi.AddProject(db, p)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not add project '%s'", p), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/projects", http.StatusSeeOther)
		return
	}
	recipt := struct {
		LoggedInUser string
	}{
		LoggedInUser: session["userid"],
	}
	err = executeTemplate(w, "add-project.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

// updateProjectHandler는 /update-project 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트 정보를 수정한다.
func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	r.ParseForm()
	id := r.Form.Get("id")
	if id == "" {
		http.Error(w, "need project 'id'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ProjectExist(db, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("project '%s' not exist", id), http.StatusBadRequest)
		return
	}
	timeForms, err := parseTimeForms(r.Form,
		"start_date",
		"release_date",
		"crank_in",
		"crank_up",
		"vfx_due_date",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Method == "POST" {
		upd := roi.UpdateProjectParam{
			Name:          r.Form.Get("name"),
			Status:        r.Form.Get("status"),
			Client:        r.Form.Get("client"),
			Director:      r.Form.Get("director"),
			Producer:      r.Form.Get("producer"),
			VFXSupervisor: r.Form.Get("vfx_supervisor"),
			VFXManager:    r.Form.Get("vfx_manager"),
			CGSupervisor:  r.Form.Get("cg_supervisor"),
			StartDate:     timeForms["start_date"],
			ReleaseDate:   timeForms["release_date"],
			CrankIn:       timeForms["crank_in"],
			CrankUp:       timeForms["crank_up"],
			VFXDueDate:    timeForms["vfx_due_date"],
			OutputSize:    r.Form.Get("output_size"),
			ViewLUT:       r.Form.Get("view_lut"),
			DefaultTasks:  fields(r.Form.Get("default_tasks"), ","),
		}
		err = roi.UpdateProject(db, id, upd)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("could not add project '%s'", id), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/projects", http.StatusSeeOther)
		return
	}
	p, err := roi.GetProject(db, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get project: %s", id), http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.Error(w, fmt.Sprintf("could not get project: %s", id), http.StatusBadRequest)
		return
	}
	recipt := struct {
		LoggedInUser string
		Project      *roi.Project
	}{
		LoggedInUser: session["userid"],
		Project:      p,
	}
	err = executeTemplate(w, "update-project.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

// searchHandler는 /search/ 하위 페이지로 사용자가 접속했을때 페이지를 반환한다.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	prj := r.URL.Path[len("/search/"):]

	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	prjRows, err := db.Query("SELECT id FROM projects")
	if err != nil {
		fmt.Fprintln(os.Stderr, "project selection error: ", err)
		return
	}
	defer prjRows.Close()
	prjs := make([]string, 0)
	for prjRows.Next() {
		p := ""
		if err := prjRows.Scan(&p); err != nil {
			fmt.Fprintln(os.Stderr, "error getting prject info from database: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		prjs = append(prjs, p)
	}

	if prj == "" && len(prjs) != 0 {
		// 할일: 추후 사용자가 마지막으로 선택했던 프로젝트로 이동
		http.Redirect(w, r, "/search/"+prjs[0], http.StatusSeeOther)
		return
	}
	found := false
	for _, p := range prjs {
		if p == prj {
			found = true
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "not found project %s\n", prj)
		return
		// http.Error(w, fmt.Sprintf("not found project: %s", id), http.StatusNotFound)
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	shotFilter := r.Form.Get("shot")
	tagFilter := r.Form.Get("tag")
	statusFilter := r.Form.Get("status")
	assigneeFilter := r.Form.Get("assignee")
	shots, err := roi.SearchShots(db, prj, shotFilter, tagFilter, statusFilter, assigneeFilter)
	if err != nil {
		log.Fatal(err)
	}
	tasks := make(map[string]map[string]*roi.Task)
	for _, s := range shots {
		ts, err := roi.AllTasks(db, prj, s.ID)
		if err != nil {
			log.Printf("could not get all tasks of shot '%s'", s.ID)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		tm := make(map[string]*roi.Task)
		for _, t := range ts {
			tm[t.Name] = t
		}
		tasks[s.ID] = tm
	}

	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}

	recipt := struct {
		LoggedInUser   string
		Projects       []string
		Project        string
		Shots          []*roi.Shot
		AllShotStatus  []roi.ShotStatus
		Tasks          map[string]map[string]*roi.Task
		FilterShot     string
		FilterTag      string
		FilterStatus   string
		FilterAssignee string
	}{
		LoggedInUser:   session["userid"],
		Projects:       prjs,
		Project:        prj,
		Shots:          shots,
		AllShotStatus:  roi.AllShotStatus,
		Tasks:          tasks,
		FilterShot:     shotFilter,
		FilterTag:      tagFilter,
		FilterStatus:   statusFilter,
		FilterAssignee: assigneeFilter,
	}
	err = executeTemplate(w, "search.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

// shotHandler는 /shot/ 하위 페이지로 사용자가 접속했을때 샷 정보가 담긴 페이지를 반환한다.
func shotHandler(w http.ResponseWriter, r *http.Request) {
	pth := r.URL.Path[len("/shot/"):]
	pths := strings.Split(pth, "/")
	if len(pths) != 2 {
		http.NotFound(w, r)
		return
	}
	prj := pths[0]
	shot := pths[1]
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}
	s, err := roi.GetShot(db, prj, shot)
	if err != nil {
		log.Fatal(err)
	}
	if s.ID == "" {
		http.NotFound(w, r)
		return
	}
	tasks, err := roi.AllTasks(db, prj, s.ID)
	if err != nil {
		log.Printf("could not get all tasks of shot '%s'", s.ID)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	recipt := struct {
		LoggedInUser string
		Project      string
		Shot         *roi.Shot
		Tasks        []*roi.Task
	}{
		LoggedInUser: session["userid"],
		Project:      prj,
		Shot:         s,
		Tasks:        tasks,
	}
	err = executeTemplate(w, "shot.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

func addShotHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	r.ParseForm()
	// 어떤 프로젝트에 샷을 생성해야 하는지 체크.
	prj := r.Form.Get("project_id")
	if prj == "" {
		// 할일: 현재 GUI 디자인으로는 프로젝트를 선택하기 어렵기 때문에
		// 일단 첫번째 프로젝트로 이동한다. 나중에는 에러가 나야 한다.
		// 관련 이슈: #143
		prjRows, err := db.Query("SELECT id FROM projects")
		if err != nil {
			log.Print("could not select the first project:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer prjRows.Close()
		if !prjRows.Next() {
			fmt.Fprintf(w, "no projects in roi yet")
			return
		}
		if err := prjRows.Scan(&prj); err != nil {
			log.Printf("could not scan a row of project '%s': %v", prj, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/add-shot/?project_id="+prj, http.StatusSeeOther)
		return
	}
	p, err := roi.GetProject(db, prj)
	if err != nil {
		log.Printf("could not get project '%s': %v", prj, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if p == nil {
		msg := fmt.Sprintf("project '%s' not exist", prj)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if r.Method == "POST" {
		shot := r.Form.Get("id")
		if shot == "" {
			http.Error(w, "need shot 'id'", http.StatusBadRequest)
			return
		}
		exist, err := roi.ShotExist(db, prj, shot)
		if err != nil {
			log.Printf("could not check shot '%s' exist", shot)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if exist {
			http.Error(w, "shot '%s' already exist", http.StatusBadRequest)
			return
		}
		tasks := fields(r.Form.Get("working_tasks"), ",")
		s := &roi.Shot{
			ID:            shot,
			ProjectID:     prj,
			Status:        roi.ShotWaiting,
			EditOrder:     atoi(r.Form.Get("edit_order")),
			Description:   r.Form.Get("description"),
			CGDescription: r.Form.Get("cg_description"),
			TimecodeIn:    r.Form.Get("timecode_in"),
			TimecodeOut:   r.Form.Get("timecode_out"),
			Duration:      atoi(r.Form.Get("duration")),
			Tags:          fields(r.Form.Get("tags"), ","),
			WorkingTasks:  tasks,
		}
		err = roi.AddShot(db, prj, s)
		if err != nil {
			log.Printf("could not add shot '%s': %v", prj+"."+shot, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		for _, task := range tasks {
			t := &roi.Task{
				ProjectID: prj,
				ShotID:    shot,
				Name:      task,
			}
			roi.AddTask(db, prj, shot, t)
		}
		http.Redirect(w, r, fmt.Sprintf("/shot/%s/%s", prj, shot), http.StatusSeeOther)
		return
	}
	recipt := struct {
		LoggedInUser string
		Project      *roi.Project
	}{
		LoggedInUser: session["userid"],
		Project:      p,
	}
	err = executeTemplate(w, "add-shot.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

func updateShotHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	r.ParseForm()
	prj := r.Form.Get("project_id")
	if prj == "" {
		http.Error(w, "need 'project_id'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("project '%s' not exist", prj), http.StatusBadRequest)
		return
	}
	shot := r.Form.Get("id")
	if shot == "" {
		http.Error(w, "need 'id'", http.StatusBadRequest)
		return
	}
	if r.Method == "POST" {
		exist, err = roi.ShotExist(db, prj, shot)
		if err != nil {
			log.Print(err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !exist {
			http.Error(w, fmt.Sprintf("shot '%s' not exist", shot), http.StatusBadRequest)
			return
		}
		tasks := fields(r.Form.Get("working_tasks"), ",")
		tforms, err := parseTimeForms(r.Form, "due_date")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		upd := roi.UpdateShotParam{
			Status:        roi.ShotStatus(r.Form.Get("status")),
			EditOrder:     atoi(r.Form.Get("edit_order")),
			Description:   r.Form.Get("description"),
			CGDescription: r.Form.Get("cg_description"),
			TimecodeIn:    r.Form.Get("timecode_in"),
			TimecodeOut:   r.Form.Get("timecode_out"),
			Duration:      atoi(r.Form.Get("duration")),
			Tags:          fields(r.Form.Get("tags"), ","),
			WorkingTasks:  tasks,
			DueDate:       tforms["due_date"],
		}
		err = roi.UpdateShot(db, prj, shot, upd)
		if err != nil {
			log.Print(err)
			http.Error(w, fmt.Sprintf("could not update shot '%s'", shot), http.StatusInternalServerError)
			return
		}
		// 샷에 등록된 태스크 중 기존에 없었던 태스크가 있다면 생성한다.
		for _, task := range tasks {
			t := &roi.Task{
				ProjectID: prj,
				ShotID:    shot,
				Name:      task,
			}
			exist, err := roi.TaskExist(db, prj, shot, task)
			if err != nil {
				log.Printf("could not check task '%s' exist: %v", prj+"."+shot+"."+task, err)
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if !exist {
				roi.AddTask(db, prj, shot, t)
			}
		}
		http.Redirect(w, r, fmt.Sprintf("/shot/%s/%s", prj, shot), http.StatusSeeOther)
		return
	}
	s, err := roi.GetShot(db, prj, shot)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if s == nil {
		http.Error(w, fmt.Sprintf("shot '%s' not exist", shot), http.StatusBadRequest)
		return
	}
	ts, err := roi.AllTasks(db, prj, shot)
	if err != nil {
		log.Printf("could not get all tasks of shot '%s': %v", prj+"."+shot, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	tm := make(map[string]*roi.Task)
	for _, t := range ts {
		tm[t.Name] = t
	}
	recipt := struct {
		LoggedInUser  string
		Shot          *roi.Shot
		AllShotStatus []roi.ShotStatus
		Tasks         map[string]*roi.Task
		AllTaskStatus []roi.TaskStatus
	}{
		LoggedInUser:  session["userid"],
		Shot:          s,
		AllShotStatus: roi.AllShotStatus,
		Tasks:         tm,
		AllTaskStatus: roi.AllTaskStatus,
	}
	err = executeTemplate(w, "update-shot.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	r.ParseForm()
	prj := r.Form.Get("project_id")
	if prj == "" {
		http.Error(w, "need 'project_id'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("project '%s' not exist", prj), http.StatusBadRequest)
		return
	}
	shot := r.Form.Get("shot_id")
	if shot == "" {
		http.Error(w, "need 'shot_id'", http.StatusBadRequest)
		return
	}
	task := r.Form.Get("name")
	if task == "" {
		http.Error(w, "need 'name'", http.StatusBadRequest)
		return
	}
	taskID := prj + "." + shot + "." + task
	if r.Method == "POST" {
		exist, err = roi.TaskExist(db, prj, shot, task)
		if err != nil {
			log.Printf("could not check task '%s' exist: %v", taskID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !exist {
			http.Error(w, fmt.Sprintf("task '%s' not exist", taskID), http.StatusBadRequest)
			return
		}
		upd := roi.UpdateTaskParam{
			Status:   roi.TaskStatus(r.Form.Get("status")),
			Assignee: r.Form.Get("assignee"),
		}
		err = roi.UpdateTask(db, prj, shot, task, upd)
		if err != nil {
			log.Printf("could not update task '%s': %v", taskID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	}
}

func addVersionHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	r.ParseForm()
	prj := r.Form.Get("project_id")
	if prj == "" {
		http.Error(w, "need 'project_id'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("project '%s' not exist", prj), http.StatusBadRequest)
		return
	}
	shot := r.Form.Get("shot_id")
	if shot == "" {
		http.Error(w, "need 'shot_id'", http.StatusBadRequest)
		return
	}
	task := r.Form.Get("task_name")
	if task == "" {
		http.Error(w, "need 'task_name'", http.StatusBadRequest)
		return
	}
	// addVersion은 새 버전을 추가하는 역할만 하고 값을 넣는 역할은 하지 않는다.
	// 만일 인수를 받았다면 에러를 낼 것.
	version := r.Form.Get("version")
	if version != "" {
		// 버전은 db에 기록된 마지막 버전을 기준으로 하지 여기서 받아들이지 않는다.
		http.Error(w, "'version' should not be specified", http.StatusBadRequest)
		return
	}
	if r.Form.Get("files") != "" {
		http.Error(w, "does not accept 'files'", http.StatusBadRequest)
		return
	}
	if r.Form.Get("mov") != "" {
		http.Error(w, "does not accept 'mov'", http.StatusBadRequest)
		return
	}
	if r.Form.Get("work_file") != "" {
		http.Error(w, "does not accept 'work_file'", http.StatusBadRequest)
		return
	}
	if r.Form.Get("created") != "" {
		http.Error(w, "does not accept 'created'", http.StatusBadRequest)
		return
	}
	taskID := fmt.Sprintf("%s.%s.%s", prj, shot, task)
	t, err := roi.GetTask(db, prj, shot, task)
	if err != nil {
		log.Printf("could not get task '%s': %v", taskID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, fmt.Sprintf("task '%s' not exist", taskID), http.StatusBadRequest)
		return
	}
	o := &roi.Version{
		ProjectID: prj,
		ShotID:    shot,
		TaskName:  task,
	}
	err = roi.AddVersion(db, prj, shot, task, o)
	if err != nil {
		log.Printf("could not add version to task '%s': %v", taskID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/search/"+prj, http.StatusSeeOther)
	return
}

func updateVersionHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	r.ParseForm()
	prj := r.Form.Get("project_id")
	if prj == "" {
		http.Error(w, "need 'project_id'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("project '%s' not exist", prj), http.StatusBadRequest)
		return
	}
	shot := r.Form.Get("shot_id")
	if shot == "" {
		http.Error(w, "need 'shot_id'", http.StatusBadRequest)
		return
	}
	task := r.Form.Get("task_name")
	if task == "" {
		http.Error(w, "need 'task_name'", http.StatusBadRequest)
		return
	}
	v := r.Form.Get("version")
	if v == "" {
		http.Error(w, "need 'version'", http.StatusBadRequest)
		return
	}
	version, err := strconv.Atoi(v)
	if err != nil {
		http.Error(w, "'version' is not a number", http.StatusBadRequest)
		return
	}
	taskID := fmt.Sprintf("%s.%s.%s", prj, shot, task)
	t, err := roi.GetTask(db, prj, shot, task)
	if err != nil {
		log.Printf("could not get task '%s': %v", taskID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, fmt.Sprintf("task '%s' not exist", taskID), http.StatusBadRequest)
		return
	}
	versionID := fmt.Sprintf("%s.%s.%s.v%v03d", prj, shot, task, version)
	if r.Method == "POST" {
		exist, err := roi.VersionExist(db, prj, shot, task, version)
		if err != nil {
			log.Printf("could not check version '%s' exist: %v", versionID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if !exist {
			http.Error(w, "version '%s' not exist", http.StatusBadRequest)
			return
		}
		timeForms, err := parseTimeForms(r.Form,
			"created",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		u := roi.UpdateVersionParam{
			OutputFiles: fields(r.Form.Get("output_files"), ","),
			Images:      fields(r.Form.Get("images"), ","),
			Mov:         r.Form.Get("mov"),
			WorkFile:    r.Form.Get("work_file"),
			Created:     timeForms["created"],
		}
		roi.UpdateVersion(db, prj, shot, task, version, u)
		http.Redirect(w, r, "/search/"+prj, http.StatusSeeOther)
		return
	}
	o, err := roi.GetVersion(db, prj, shot, task, version)
	if err != nil {
		log.Printf("could not get version '%s': %v", versionID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if o == nil {
		http.Error(w, fmt.Sprintf("version '%s' not exist", versionID), http.StatusBadRequest)
		return
	}
	recipt := struct {
		LoggedInUser string
		Version      *roi.Version
	}{
		LoggedInUser: session["userid"],
		Version:      o,
	}
	err = executeTemplate(w, "update-version.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func main() {
	dev = true

	var (
		init  bool
		https string
		cert  string
		key   string
	)
	flag.BoolVar(&init, "init", false, "setup roi.")
	flag.StringVar(&https, "https", ":443", "address to open https port. it doesn't offer http for security reason.")
	flag.StringVar(&cert, "cert", "cert/cert.pem", "https cert file. default one for testing will created by -init.")
	flag.StringVar(&key, "key", "cert/key.pem", "https key file. default one for testing will created by -init.")
	flag.Parse()

	hashFile := "cert/cookie.hash"
	blockFile := "cert/cookie.block"

	if init {
		// 기본 Self Signed Certificate는 항상 정해진 위치에 생성되어야 한다.
		cert := "cert/cert.pem"
		key := "cert/key.pem"
		// 해당 위치에 이미 파일이 생성되어 있다면 건너 뛴다.
		// 사용자가 직접 추가한 인증서 파일을 덮어쓰는 위험을 없애기 위함이다.
		exist, err := anyFileExist(cert, key)
		if err != nil {
			log.Fatalf("error checking a certificate file %s: %s", cert, err)
		}
		if exist {
			log.Print("already have certificate file. will not create.")
		} else {
			// cert와 key가 없다. 인증서 생성.
			c := exec.Command("sh", "generate-self-signed-cert.sh")
			c.Dir = "cert"
			_, err := c.CombinedOutput()
			if err != nil {
				log.Fatal("error generating certificate files: ", err)
			}
		}

		exist, err = anyFileExist(hashFile, blockFile)
		if err != nil {
			log.Fatalf("could not check cookie key file: %s", err)
		}
		if exist {
			log.Print("already have cookie file. will not create.")
		} else {
			ioutil.WriteFile(hashFile, securecookie.GenerateRandomKey(64), 0600)
			ioutil.WriteFile(blockFile, securecookie.GenerateRandomKey(32), 0600)
		}
		return
	}

	err := roi.InitDB()
	if err != nil {
		log.Fatalf("could not initialize database: %v", err)
	}

	parseTemplate()

	hashKey, err := ioutil.ReadFile(hashFile)
	if err != nil {
		log.Fatalf("could not read cookie hash key from file '%s'", hashFile)
	}
	blockKey, err := ioutil.ReadFile(blockFile)
	if err != nil {
		log.Fatalf("could not read cookie block key from file '%s'", blockFile)
	}
	cookieHandler = securecookie.New(
		hashKey,
		blockKey,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/login/", loginHandler)
	mux.HandleFunc("/logout/", logoutHandler)
	mux.HandleFunc("/settings/profile", profileHandler)
	mux.HandleFunc("/update-password", updatePasswordHandler)
	mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/projects", projectsHandler)
	mux.HandleFunc("/add-project", addProjectHandler)
	mux.HandleFunc("/update-project", updateProjectHandler)
	mux.HandleFunc("/search/", searchHandler)
	mux.HandleFunc("/shot/", shotHandler)
	mux.HandleFunc("/add-shot/", addShotHandler)
	mux.HandleFunc("/update-shot", updateShotHandler)
	mux.HandleFunc("/update-task", updateTaskHandler)
	mux.HandleFunc("/add-version", addVersionHandler)
	mux.HandleFunc("/update-version", updateVersionHandler)
	mux.HandleFunc("/api/v1/shot/add", addShotApiHandler)
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	thumbfs := http.FileServer(http.Dir("roi-userdata/thumbnail"))
	mux.Handle("/thumbnail/", http.StripPrefix("/thumbnail/", thumbfs))

	// Show https binding information
	addrToShow := "https://"
	addrs := strings.Split(https, ":")
	if len(addrs) == 2 {
		if addrs[0] == "" {
			addrToShow += "localhost"
		} else {
			addrToShow += addrs[0]
		}
		if addrs[1] != "443" {
			addrToShow += ":" + addrs[1]
		}
	}
	fmt.Println()
	log.Printf("roi is start to running. see %s", addrToShow)
	fmt.Println()

	// Bind
	log.Fatal(http.ListenAndServeTLS(https, cert, key, mux))
}
