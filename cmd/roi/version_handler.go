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
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, ctg, unit, task, err := roi.SplitTaskID(id)
	if err != nil {
		return err
	}
	recipe := struct {
		PageType     string
		LoggedInUser string
		Version      *roi.Version
	}{
		LoggedInUser: env.User.ID,
		Version: &roi.Version{
			Show:     show,
			Category: ctg,
			Unit:     unit,
			Task:     task,
		},
	}
	return executeTemplate(w, "add-version.html", recipe)
}

func addVersionPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id", "version")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, ctg, unit, task, err := roi.SplitTaskID(id)
	if err != nil {
		return err
	}
	version := r.FormValue("version")
	v := &roi.Version{
		Show:      show,
		Category:  ctg,
		Unit:      unit,
		Task:      task,
		Status:    roi.VersionInProgress,
		Version:   version,
		StartDate: time.Now(),
		Owner:     env.User.ID,
	}
	err = roi.AddVersion(DB, v)
	if err != nil {
		return err
	}
	http.Redirect(w, r, fmt.Sprintf("/update-version?id=%s", id+"/"+version), http.StatusSeeOther)
	return nil
}

func updateVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateVersionPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	v, err := roi.GetVersion(DB, id)
	if err != nil {
		return err
	}
	t, err := roi.GetTask(DB, v.TaskID())
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser     string
		Version          *roi.Version
		IsWorkingVersion bool
		IsPublishVersion bool
		AllVersionStatus []roi.VersionStatus
	}{
		LoggedInUser:     env.User.ID,
		Version:          v,
		IsWorkingVersion: t.WorkingVersion == v.Version,
		IsPublishVersion: t.PublishVersion == v.Version,
		AllVersionStatus: roi.AllVersionStatus,
	}
	return executeTemplate(w, "update-version.html", recipe)
}

func updateVersionPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	timeForms, err := parseTimeForms(r.Form, "start_date", "end_date")
	if err != nil {
		return err
	}
	mov := fmt.Sprintf("data/show/%s/1.mov", id)
	err = saveFormFile(r, "mov", mov)
	if err != nil {
		return err
	}
	v, err := roi.GetVersion(DB, id)
	if err != nil {
		return err
	}
	v.OutputFiles = fieldSplit(r.FormValue("output_files"))
	v.Images = fieldSplit(r.FormValue("images"))
	v.WorkFile = r.FormValue("work_file")
	v.StartDate = timeForms["start_date"]
	v.EndDate = timeForms["end_date"]

	err = roi.UpdateVersion(DB, id, v)
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
	err := mustFields(r, "id", "update-status")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	status := roi.VersionStatus(r.FormValue("update-status"))
	err = roi.UpdateVersionStatus(DB, id, status)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
