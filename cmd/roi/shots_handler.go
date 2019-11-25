package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// shotsHandler는 /shots/ 페이지로 사용자가 접속했을때 페이지를 반환한다.
func shotsHandler(w http.ResponseWriter, r *http.Request) {
	show := r.URL.Path[len("/shots/"):]

	ps, err := roi.AllShows(DB)
	if err != nil {
		log.Printf("could not get show list: %v", err)
	}
	shows := make([]string, len(ps))
	for i, p := range ps {
		shows[i] = p.Show
	}
	if show == "" && len(shows) != 0 {
		// 할일: 추후 사용자가 마지막으로 선택했던 프로젝트로 이동
		http.Redirect(w, r, "/shots/"+shows[0], http.StatusSeeOther)
		return
	}
	if show != "" {
		found := false
		for _, p := range shows {
			if p == show {
				found = true
				break
			}
		}
		if !found {
			http.Error(w, "show not found", http.StatusNotFound)
			return
		}
	}

	if show == "" {
		// TODO: show empty page
		// for now SearchShot will handle it properly, don't panic.
	}

	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	query := r.Form.Get("q")
	f := make(map[string]string)
	for _, v := range strings.Fields(query) {
		kv := strings.Split(v, ":")
		if len(kv) == 1 {
			f["shot"] = v
		} else {
			f[kv[0]] = kv[1]
		}
	}
	// toTime은 문자열을 time.Time으로 변경하되 에러가 나면 버린다.
	toTime := func(s string) time.Time {
		t, _ := timeFromString(s)
		return t
	}
	shots, err := roi.SearchShots(DB, show, f["shot"], f["tag"], f["status"], f["assignee"], f["task-status"], toTime(f["task-due"]))
	if err != nil {
		log.Fatal(err)
	}
	tasks := make(map[string]map[string]*roi.Task)
	for _, s := range shots {
		ts, err := roi.ShotTasks(DB, show, s.Shot)
		if err != nil {
			log.Printf("couldn't get shot tasks: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		tm := make(map[string]*roi.Task)
		for _, t := range ts {
			tm[t.Task] = t
		}
		tasks[s.Shot] = tm
	}

	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}

	recipe := struct {
		LoggedInUser  string
		Shows         []string
		Show          string
		Shots         []*roi.Shot
		AllShotStatus []roi.ShotStatus
		Tasks         map[string]map[string]*roi.Task
		AllTaskStatus []roi.TaskStatus
		Query         string
	}{
		LoggedInUser:  session["userid"],
		Shows:         shows,
		Show:          show,
		Shots:         shots,
		AllShotStatus: roi.AllShotStatus,
		Tasks:         tasks,
		AllTaskStatus: roi.AllTaskStatus,
		Query:         query,
	}
	err = executeTemplate(w, "shots.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}
