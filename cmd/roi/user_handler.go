package main

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// loginHandler는 /login 페이지로 사용자가 접속했을때 로그인 페이지를 반환한다.
func loginHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		err := mustFields(r, "id", "password")
		if err != nil {
			return err
		}
		id := r.FormValue("id")
		pw := r.FormValue("password")
		match, err := roi.UserPasswordMatch(DB, id, pw)
		if err != nil {
			return err
		}
		if !match {
			return roi.BadRequest("entered password is not correct")
		}
		session := map[string]string{
			"userid": id,
		}
		err = setSession(w, session)
		if err != nil {
			return fmt.Errorf("could not set session: %w", err)
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	return executeTemplate(w, "login.bml", nil)
}

// logoutHandler는 /logout 페이지로 사용자가 접속했을때 사용자를 로그아웃 시킨다.
func logoutHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	clearSession(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}

// signupHandler는 /signup 페이지로 사용자가 접속했을때 가입 페이지를 반환한다.
func signupHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		err := mustFields(r, "id", "password")
		if err != nil {
			return err
		}
		id := r.FormValue("id")
		pw := r.FormValue("password")
		if len(pw) < 8 {
			return roi.BadRequest("password too short")
		}
		// 할일: password에 대한 컨펌은 프론트 엔드에서 하여야 함
		pwc := r.FormValue("password_confirm")
		if pw != pwc {
			return roi.BadRequest("passwords are not matched")
		}
		err = roi.AddUser(DB, id, pw)
		if err != nil {
			return err
		}
		session := map[string]string{
			"userid": id,
		}
		err = setSession(w, session)
		if err != nil {
			return err
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil
	}
	return executeTemplate(w, "signup.bml", nil)
}

// profileHandler는 /profile 페이지로 사용자가 접속했을 때 사용자 프로필 페이지를 반환한다.
func profileHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		id := r.FormValue("id")
		if env.User.ID != id {
			return roi.BadRequest("not allowed to change other's profile")
		}
		u, err := roi.GetUser(DB, id)
		if err != nil {
			return err
		}
		u.KorName = r.FormValue("kor_name")
		u.Name = r.FormValue("name")
		u.Team = r.FormValue("team")
		u.Role = r.FormValue("position")
		u.Email = r.FormValue("email")
		u.PhoneNumber = r.FormValue("phone_number")
		u.EntryDate = r.FormValue("entry_date")

		err = roi.UpdateUser(DB, id, u)
		if err != nil {
			return err
		}
		http.Redirect(w, r, "/settings/profile", http.StatusSeeOther)
		return nil
	}
	recipe := struct {
		LoggedInUser string
		User         *roi.User
	}{
		LoggedInUser: env.User.ID,
		User:         env.User,
	}
	return executeTemplate(w, "profile.bml", recipe)
}

// updatePasswordHandler는 /update-password 페이지로 사용자가 패스워드 변경과 관련된 정보를 보내면
// 사용자 패스워드를 변경한다.
func updatePasswordHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "old_password", "new_password")
	if err != nil {
		return err
	}
	oldpw := r.FormValue("old_password")
	newpw := r.FormValue("new_password")
	if len(newpw) < 8 {
		return roi.BadRequest("new password too short")
	}
	// 할일: password에 대한 컨펌은 프론트 엔드에서 하여야 함
	newpwc := r.FormValue("new_password_confirm")
	if newpw != newpwc {
		return roi.BadRequest("passwords are not matched")
	}
	match, err := roi.UserPasswordMatch(DB, env.User.ID, oldpw)
	if err != nil {
		return err
	}
	if !match {
		return roi.BadRequest("entered password is not correct")
	}
	err = roi.UpdateUserPassword(DB, env.User.ID, newpw)
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/settings/profile", http.StatusSeeOther)
	return nil
}

// userHandler는 루트 페이지로 사용자가 접근했을때 그 사용자에게 필요한 정보를 맞춤식으로 제공한다.
func userHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	user := r.URL.Path[len("/user/"):]
	_, err := roi.GetUser(DB, user)
	if err != nil {
		return err
	}
	tasks, err := roi.UserTasks(DB, user)
	if err != nil {
		return err
	}
	// 태스크를 미리 아이디 기준으로 정렬해 두면 아래에서 사용되는
	// tasksOfDay 또한 아이디 기준으로 정렬된다.
	sort.Slice(tasks, func(i, j int) bool {
		ti := tasks[i]
		tj := tasks[j]
		c := strings.Compare(ti.Show, tj.Show)
		if c < 0 {
			return true
		} else if c > 0 {
			return false
		}
		c = strings.Compare(ti.Category, tj.Category)
		if c < 0 {
			return true
		} else if c > 0 {
			return false
		}
		c = strings.Compare(ti.Unit, tj.Unit)
		if c < 0 {
			return true
		} else if c > 0 {
			return false
		}
		c = strings.Compare(ti.Task, tj.Task)
		if c <= 0 {
			return true
		}
		return false
	})
	taskFromID := make(map[string]*roi.Task)
	for _, t := range tasks {
		taskFromID[t.ID()] = t
	}
	tasksOfDay := make(map[string][]string, 28)
	for _, t := range tasks {
		due := stringFromDate(t.DueDate)
		if tasksOfDay[due] == nil {
			tasksOfDay[due] = make([]string, 0)
		}
		tasksOfDay[due] = append(tasksOfDay[due], t.ID())
	}
	// 앞으로 4주에 대한 태스크 정보를 보인다.
	// 총 기간이나 단위는 추후 설정할 수 있도록 할 것.
	timeline := make([]string, 28)
	y, m, d := time.Now().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.Local)
	for i := range timeline {
		timeline[i] = stringFromDate(today.Add(time.Duration(i) * 24 * time.Hour))
	}
	numTasks := make(map[string]map[roi.Status]int)
	for _, t := range tasks {
		if numTasks[t.Show] == nil {
			numTasks[t.Show] = make(map[roi.Status]int)
		}
		numTasks[t.Show][t.Status] += 1
	}
	recipe := struct {
		LoggedInUser  string
		User          string
		Timeline      []string
		NumTasks      map[string]map[roi.Status]int
		TaskFromID    map[string]*roi.Task
		TasksOfDay    map[string][]string
		AllTaskStatus []roi.Status
	}{
		LoggedInUser:  env.User.ID,
		User:          user,
		Timeline:      timeline,
		NumTasks:      numTasks,
		TaskFromID:    taskFromID,
		TasksOfDay:    tasksOfDay,
		AllTaskStatus: roi.AllTaskStatus,
	}
	return executeTemplate(w, "user.bml", recipe)
}

func usersHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	us, err := roi.Users(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser string
		Users        []*roi.User
	}{
		LoggedInUser: env.User.ID,
		Users:        us,
	}
	return executeTemplate(w, "users.bml", recipe)
}
