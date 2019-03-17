package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/studio2l/roi"
)

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
	taskStatusFilter := r.Form.Get("task_status")
	tforms, err := parseTimeForms(r.Form, "task_due_date")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	taskDueDateFilter := tforms["task_due_date"]
	shots, err := roi.SearchShots(db, prj, shotFilter, tagFilter, statusFilter, assigneeFilter, taskStatusFilter, taskDueDateFilter)
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
		LoggedInUser      string
		Projects          []string
		Project           string
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
		Projects:          prjs,
		Project:           prj,
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
	err = executeTemplate(w, "search.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}
