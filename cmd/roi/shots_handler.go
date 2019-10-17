package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/studio2l/roi"
)

// shotsHandler는 /shots/ 페이지로 사용자가 접속했을때 페이지를 반환한다.
func shotsHandler(w http.ResponseWriter, r *http.Request) {
	show := r.URL.Path[len("/shots/"):]

	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	ps, err := roi.AllShows(db)
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
	shotFilter := r.Form.Get("shot")
	tagFilter := r.Form.Get("tag")
	statusFilter := r.Form.Get("status")
	assigneeFilter := r.Form.Get("assignee")
	taskStatusFilter := r.Form.Get("task_status")
	tforms, err := parseTimeForms(r.Form, "task_due_date")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	taskDueDateFilter := tforms["task_due_date"]
	shots, err := roi.SearchShots(db, show, shotFilter, tagFilter, statusFilter, assigneeFilter, taskStatusFilter, taskDueDateFilter)
	if err != nil {
		log.Fatal(err)
	}
	tasks := make(map[string]map[string]*roi.Task)
	for _, s := range shots {
		ts, err := roi.ShotTasks(db, show, s.Shot)
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

	recipt := struct {
		LoggedInUser      string
		Shows             []string
		Show              string
		Shots             []*roi.Shot
		AllShotStatus     []roi.ShotStatus
		Tasks             map[string]map[string]*roi.Task
		AllTaskStatus     []roi.TaskStatus
		FilterShot        string
		FilterTag         string
		FilterStatus      string
		FilterAssignee    string
		FilterTaskStatus  string
		FilterTaskDueDate time.Time
	}{
		LoggedInUser:      session["userid"],
		Shows:             shows,
		Show:              show,
		Shots:             shots,
		AllShotStatus:     roi.AllShotStatus,
		Tasks:             tasks,
		AllTaskStatus:     roi.AllTaskStatus,
		FilterShot:        shotFilter,
		FilterTag:         tagFilter,
		FilterStatus:      statusFilter,
		FilterAssignee:    assigneeFilter,
		FilterTaskStatus:  taskStatusFilter,
		FilterTaskDueDate: taskDueDateFilter,
	}
	err = executeTemplate(w, "shots.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}
