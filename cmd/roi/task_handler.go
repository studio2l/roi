package main

import (
	"net/http"

	"github.com/studio2l/roi"
)

func updateTaskHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateTaskPostHandler(w, r, env)
	}
	err := mustFields(r, "show", "shot", "task")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	t, err := roi.GetTask(DB, show+"/"+shot+"/"+task)
	if err != nil {
		return err
	}
	vers, err := roi.TaskVersions(DB, show+"/"+shot+"/"+task)
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
	err := mustFields(r, "show", "shot", "task")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
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
	err = roi.UpdateTask(DB, show+"/"+shot+"/"+task, upd)
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
	err := mustFields(r, "show", "shot", "task", "version")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	version := r.FormValue("version")
	err = roi.UpdateTaskWorkingVersion(DB, show+"/"+shot+"/"+task, version)
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
	err := mustFields(r, "show", "shot", "task", "version")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	version := r.FormValue("version")
	err = roi.UpdateTaskPublishVersion(DB, show+"/"+shot+"/"+task, version)
	if err != nil {
		return err
	}
	// 수정 페이지로 돌아간다.
	r.Method = "GET"
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
