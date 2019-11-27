package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/studio2l/roi"
)

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	u, err := sessionUser(r)
	if err != nil {
		handleError(w, err)
		clearSession(w)
		return
	}
	if u == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	show := r.FormValue("show")
	if show == "" {
		http.Error(w, "need 'show'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ShowExist(DB, show)
	if err != nil {
		log.Print(err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("show '%s' not exist", show), http.StatusBadRequest)
		return
	}
	shot := r.FormValue("shot")
	if shot == "" {
		http.Error(w, "need 'shot'", http.StatusBadRequest)
		return
	}
	task := r.FormValue("task")
	if task == "" {
		http.Error(w, "need 'task'", http.StatusBadRequest)
		return
	}
	taskID := show + "." + shot + "." + task
	if r.Method == "POST" {
		exist, err = roi.TaskExist(DB, show, shot, task)
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
			Status:   roi.TaskStatus(r.FormValue("status")),
			Assignee: r.FormValue("assignee"),
			DueDate:  tforms["due_date"],
		}
		err = roi.UpdateTask(DB, show, shot, task, upd)
		if err != nil {
			log.Printf("could not update task '%s': %v", taskID, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// 수정 페이지로 돌아간다.
		r.Method = "GET"
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	t, err := roi.GetTask(DB, show, shot, task)
	if err != nil {
		log.Printf("could not get task '%s': %v", taskID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if t == nil {
		http.Error(w, fmt.Sprintf("task '%s' not exist", taskID), http.StatusBadRequest)
		return
	}
	vers, err := roi.TaskVersions(DB, show, shot, task)
	if err != nil {
		log.Printf("could not get versions of '%s': %v", taskID, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	recipe := struct {
		LoggedInUser  string
		Task          *roi.Task
		AllTaskStatus []roi.TaskStatus
		Versions      []*roi.Version // 역순
	}{
		LoggedInUser:  u.ID,
		Task:          t,
		AllTaskStatus: roi.AllTaskStatus,
		Versions:      vers,
	}
	err = executeTemplate(w, "update-task.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}
