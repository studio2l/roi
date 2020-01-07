package main

import (
	"net/http"

	"github.com/studio2l/roi"
)

func updateTaskHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateTaskPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	err = roi.VerifyTaskID(id)
	if err != nil {
		return err
	}
	t, err := roi.GetTask(DB, id)
	if err != nil {
		return err
	}
	vers, err := roi.TaskVersions(DB, id)
	if err != nil {
		return err
	}
	us, err := roi.Users(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser  string
		Task          *roi.Task
		AllTaskStatus []roi.TaskStatus
		Versions      []*roi.Version
		Users         []*roi.User
	}{
		LoggedInUser:  env.SessionUser.ID,
		Task:          t,
		AllTaskStatus: roi.AllTaskStatus,
		Versions:      vers,
		Users:         us,
	}
	return executeTemplate(w, "update-task.html", recipe)
}

func updateTaskPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	err = roi.VerifyTaskID(id)
	if err != nil {
		return err
	}
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
	t, err := roi.GetTask(DB, id)
	if err != nil {
		return err
	}
	t.Status = roi.TaskStatus(r.FormValue("status"))
	t.Assignee = assignee
	t.DueDate = tforms["due_date"]

	err = roi.UpdateTask(DB, id, t)
	if err != nil {
		return err
	}
	// 수정 페이지로 돌아간다.
	r.Method = "GET"
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateTaskWorkingVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method != "POST" {
		return roi.BadRequest("only post method allowed")
	}
	err := mustFields(r, "id", "version")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	err = roi.VerifyTaskID(id)
	if err != nil {
		return err
	}
	version := r.FormValue("version")
	err = roi.UpdateTaskWorkingVersion(DB, id, version)
	if err != nil {
		return err
	}
	// 수정 페이지로 돌아간다.
	r.Method = "GET"
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateTaskPublishVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method != "POST" {
		return roi.BadRequest("only post method allowed")
	}
	err := mustFields(r, "id", "version")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	err = roi.VerifyTaskID(id)
	if err != nil {
		return err
	}
	version := r.FormValue("version")
	err = roi.UpdateTaskPublishVersion(DB, id, version)
	if err != nil {
		return err
	}
	// 수정 페이지로 돌아간다.
	r.Method = "GET"
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
