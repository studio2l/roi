package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/studio2l/roi"
)

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
	prj := r.Form.Get("project")
	if prj == "" {
		http.Error(w, "need 'project'", http.StatusBadRequest)
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
	shot := r.Form.Get("shot")
	if shot == "" {
		http.Error(w, "need 'shot'", http.StatusBadRequest)
		return
	}
	task := r.Form.Get("task")
	if task == "" {
		http.Error(w, "need 'task'", http.StatusBadRequest)
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
		tforms, err := parseTimeForms(r.Form, "due_date")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		upd := roi.UpdateTaskParam{
			Status:   roi.TaskStatus(r.Form.Get("status")),
			Assignee: r.Form.Get("assignee"),
			DueDate:  tforms["due_date"],
		}
		err = roi.UpdateTask(db, prj, shot, task, upd)
		if err != nil {
			log.Printf("could not update task '%s': %v", taskID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// 수정 페이지로 돌아간다.
		r.Method = "GET"
		http.Redirect(w, r, r.RequestURI, http.StatusSeeOther)
		return
	}
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
	vers := make([]int, t.LastOutputVersion)
	for i := range vers {
		vers[i] = t.LastOutputVersion - i
	}
	recipt := struct {
		LoggedInUser  string
		Task          *roi.Task
		AllTaskStatus []roi.TaskStatus
		Versions      []int // 역순
	}{
		LoggedInUser:  session["userid"],
		Task:          t,
		AllTaskStatus: roi.AllTaskStatus,
		Versions:      vers,
	}
	err = executeTemplate(w, "update-task.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}
