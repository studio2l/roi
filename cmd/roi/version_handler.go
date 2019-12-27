package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/studio2l/roi"
)

func addVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return addVersionPostHandler(w, r, env)
	}
	err := mustFields(r, "show", "shot", "task")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	recipe := struct {
		PageType     string
		LoggedInUser string
		Version      *roi.Version
	}{
		LoggedInUser: env.SessionUser.ID,
		Version: &roi.Version{
			Show: show,
			Shot: shot,
			Task: task,
		},
	}
	return executeTemplate(w, "add-version.html", recipe)
}

func addVersionPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show", "shot", "task", "version")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	version := r.FormValue("version")
	v := &roi.Version{
		Show:      show,
		Shot:      shot,
		Task:      task,
		Status:    roi.VersionInProgress,
		Version:   version,
		StartDate: time.Now(),
		Owner:     env.SessionUser.ID,
	}
	err = roi.AddVersion(DB, show, shot, task, v)
	if err != nil {
		return err
	}
	http.Redirect(w, r, fmt.Sprintf("/update-version?show=%s&shot=%s&task=%s&version=%s", show, shot, task, version), http.StatusSeeOther)
	return nil
}

func updateVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateVersionPostHandler(w, r, env)
	}
	err := mustFields(r, "show", "shot", "task", "version")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	version := r.FormValue("version")
	t, err := roi.GetTask(DB, show, shot, task)
	if err != nil {
		return err
	}
	v, err := roi.GetVersion(DB, show, shot, task, version)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser     string
		Version          *roi.Version
		IsWorkingVersion bool
		AllVersionStatus []roi.VersionStatus
	}{
		LoggedInUser:     env.SessionUser.ID,
		Version:          v,
		IsWorkingVersion: t.WorkingVersion == v.Version,
		AllVersionStatus: roi.AllVersionStatus,
	}
	return executeTemplate(w, "update-version.html", recipe)
}

func updateVersionPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show", "shot", "task", "version", "status")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	version := r.FormValue("version")
	timeForms, err := parseTimeForms(r.Form, "start_date", "end_date")
	if err != nil {
		return err
	}
	mov := fmt.Sprintf("data/show/%s/%s/%s/%s/1.mov", show, shot, task, version)
	err = saveFormFile(r, "mov", mov)
	if err != nil {
		return err
	}
	u := roi.UpdateVersionParam{
		OutputFiles: fieldSplit(r.FormValue("output_files")),
		Images:      fieldSplit(r.FormValue("images")),
		WorkFile:    r.FormValue("work_file"),
		StartDate:   timeForms["start_date"],
		EndDate:     timeForms["end_date"],
	}
	err = roi.UpdateVersion(DB, show, shot, task, version, u)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}

func updateVersionStatusHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method != "POST" {
		return roi.BadRequest("only post method allowed")
	}
	err := mustFields(r, "show", "shot", "task", "version", "update-status")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	version := r.FormValue("version")
	status := roi.VersionStatus(r.FormValue("update-status"))
	err = roi.UpdateVersionStatus(DB, show, shot, task, version, status)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
