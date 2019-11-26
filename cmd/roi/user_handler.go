package main

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// loginHandler는 /login 페이지로 사용자가 접속했을때 로그인 페이지를 반환한다.
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		id := r.FormValue("id")
		if id == "" {
			http.Error(w, "id field emtpy", http.StatusBadRequest)
			return
		}
		pw := r.FormValue("password")
		if pw == "" {
			http.Error(w, "password field emtpy", http.StatusBadRequest)
			return
		}
		match, err := roi.UserPasswordMatch(DB, id, pw)
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
	recipe := struct {
		LoggedInUser string
	}{
		LoggedInUser: session["userid"],
	}
	err = executeTemplate(w, "login.html", recipe)
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
		id := r.FormValue("id")
		if id == "" {
			http.Error(w, "id field emtpy", http.StatusBadRequest)
			return
		}
		pw := r.FormValue("password")
		if pw == "" {
			http.Error(w, "password field emtpy", http.StatusBadRequest)
			return
		}
		if len(pw) < 8 {
			http.Error(w, "password too short", http.StatusBadRequest)
			return
		}
		// 할일: password에 대한 컨펌은 프론트 엔드에서 하여야 함
		pwc := r.FormValue("password_confirm")
		if pw != pwc {
			http.Error(w, "passwords are not matched", http.StatusBadRequest)
			return
		}
		err := roi.AddUser(DB, id, pw)
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
	recipe := struct {
		LoggedInUser string
	}{
		LoggedInUser: session["userid"],
	}
	err = executeTemplate(w, "signup.html", recipe)
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
	if r.Method == "POST" {
		upd := roi.UpdateUserParam{
			KorName:     r.FormValue("kor_name"),
			Name:        r.FormValue("name"),
			Team:        r.FormValue("team"),
			Role:        r.FormValue("position"),
			Email:       r.FormValue("email"),
			PhoneNumber: r.FormValue("phone_number"),
			EntryDate:   r.FormValue("entry_date"),
		}
		err = roi.UpdateUser(DB, session["userid"], upd)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not set user: %s", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings/profile", http.StatusSeeOther)
		return
	}
	u, err := roi.GetUser(DB, session["userid"])
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get user: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	fmt.Println(u)
	recipe := struct {
		LoggedInUser string
		User         *roi.User
	}{
		LoggedInUser: session["userid"],
		User:         u,
	}
	err = executeTemplate(w, "profile.html", recipe)
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
	oldpw := r.FormValue("old_password")
	if oldpw == "" {
		http.Error(w, "old password field emtpy", http.StatusBadRequest)
		return
	}
	newpw := r.FormValue("new_password")
	if newpw == "" {
		http.Error(w, "new password field emtpy", http.StatusBadRequest)
		return
	}
	if len(newpw) < 8 {
		http.Error(w, "new password too short", http.StatusBadRequest)
		return
	}
	// 할일: password에 대한 컨펌은 프론트 엔드에서 하여야 함
	newpwc := r.FormValue("new_password_confirm")
	if newpw != newpwc {
		http.Error(w, "passwords are not matched", http.StatusBadRequest)
		return
	}
	id := session["userid"]
	match, err := roi.UserPasswordMatch(DB, id, oldpw)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !match {
		http.Error(w, "entered password is not correct", http.StatusBadRequest)
		return
	}
	err = roi.UpdateUserPassword(DB, id, newpw)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not change user password: %s", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/settings/profile", http.StatusSeeOther)
}

// userHandler는 루트 페이지로 사용자가 접근했을때 그 사용자에게 필요한 정보를 맞춤식으로 제공한다.
func userHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		log.Printf("could not get session: %s", err)
		clearSession(w)
	}
	user := r.URL.Path[len("/user/"):]
	tasks, err := roi.UserTasks(DB, user)
	if err != nil {
		log.Printf("could not get user tasks: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	// 태스크를 미리 아이디 기준으로 정렬해 두면 아래에서 사용되는
	// tasksOfDay 또한 아이디 기준으로 정렬된다.
	sort.Slice(tasks, func(i, j int) bool {
		ti := tasks[i]
		idi := ti.Show + "." + ti.Shot + "." + ti.Task
		tj := tasks[j]
		idj := tj.Show + "." + tj.Shot + "." + tj.Task
		return strings.Compare(idi, idj) <= 0
	})
	taskFromID := make(map[string]*roi.Task)
	for _, t := range tasks {
		tid := t.Show + "." + t.Shot + "." + t.Task
		taskFromID[tid] = t
	}
	tasksOfDay := make(map[string][]string, 28)
	for _, t := range tasks {
		due := stringFromDate(t.DueDate)
		if tasksOfDay[due] == nil {
			tasksOfDay[due] = make([]string, 0)
		}
		tid := t.Show + "." + t.Shot + "." + t.Task
		tasksOfDay[due] = append(tasksOfDay[due], tid)
	}
	// 앞으로 4주에 대한 태스크 정보를 보인다.
	// 총 기간이나 단위는 추후 설정할 수 있도록 할 것.
	timeline := make([]string, 28)
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	for i := range timeline {
		timeline[i] = stringFromDate(today.Add(time.Duration(i) * 24 * time.Hour))
	}
	numTasks := make(map[string]map[roi.TaskStatus]int)
	for _, t := range tasks {
		if numTasks[t.Show] == nil {
			numTasks[t.Show] = make(map[roi.TaskStatus]int)
		}
		numTasks[t.Show][t.Status] += 1
	}
	recipe := struct {
		LoggedInUser  string
		User          string
		Timeline      []string
		NumTasks      map[string]map[roi.TaskStatus]int
		TaskFromID    map[string]*roi.Task
		TasksOfDay    map[string][]string
		AllTaskStatus []roi.TaskStatus
	}{
		LoggedInUser:  session["userid"],
		User:          user,
		Timeline:      timeline,
		NumTasks:      numTasks,
		TaskFromID:    taskFromID,
		TasksOfDay:    tasksOfDay,
		AllTaskStatus: roi.AllTaskStatus,
	}
	err = executeTemplate(w, "user.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		log.Printf("could not get session: %s", err)
		clearSession(w)
	}
	us, err := roi.Users(DB)
	if err != nil {
		log.Printf("could not get users: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	recipe := struct {
		LoggedInUser string
		Users        []*roi.User
	}{
		LoggedInUser: session["userid"],
		Users:        us,
	}
	err = executeTemplate(w, "users.html", recipe)
	if err != nil {
		log.Fatal(err)
	}

}
