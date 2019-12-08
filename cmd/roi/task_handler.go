package main

import (
	"net/http"

	"github.com/studio2l/roi"
)

func updateTaskHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show", "shot", "task")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	t, err := roi.GetTask(DB, show, shot, task)
	if err != nil {
		return err
	}
	if r.Method == "POST" {
		tforms, err := parseTimeForms(r.Form, "due_date")
		if err != nil {
			return err
		}
		assignee := r.FormValue("assignee")
		if assignee != "" {
			_, err = roi.GetUser(DB, assignee)
			if err != nil {
				return err
			}
		}
		upd := roi.UpdateTaskParam{
			Status:   roi.TaskStatus(r.FormValue("status")),
			Assignee: assignee,
			DueDate:  tforms["due_date"],
		}
		err = roi.UpdateTask(DB, show, shot, task, upd)
		if err != nil {
			return err
		}
		// 수정 페이지로 돌아간다.
		r.Method = "GET"
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return nil
	}
	vers, err := roi.TaskVersions(DB, show, shot, task)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser  string
		Task          *roi.Task
		AllTaskStatus []roi.TaskStatus
		Versions      []*roi.Version // 역순
	}{
		LoggedInUser:  env.SessionUser.ID,
		Task:          t,
		AllTaskStatus: roi.AllTaskStatus,
		Versions:      vers,
	}
	return executeTemplate(w, "update-task.html", recipe)
}
