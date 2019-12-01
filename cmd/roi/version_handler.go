package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/studio2l/roi"
)

func addVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show", "shot", "task")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	if r.Method == "POST" {
		err := mustFields(r, "version")
		if err != nil {
			return err
		}
		version := r.FormValue("version")
		timeForms, err := parseTimeForms(r.Form, "created")
		if err != nil {
			return err
		}
		_, err = roi.GetTask(DB, show, shot, task)
		if err != nil {
			return err
		}
		mov := fmt.Sprintf("data/show/%s/%s/%s/%s/1.mov", show, shot, task, version)
		err = saveFormFile(r, "mov", mov)
		if err != nil {
			return err
		}
		v := &roi.Version{
			Show:        show,
			Shot:        shot,
			Task:        task,
			Version:     version,
			OutputFiles: fields(r.FormValue("output_files")),
			WorkFile:    r.FormValue("work_file"),
			Created:     timeForms["create"],
		}
		err = roi.AddVersion(DB, show, shot, task, v)
		if err != nil {
			return err
		}
		http.Redirect(w, r, fmt.Sprintf("/update-version?show=%s&shot=%s&task=%s&version=%s", show, shot, task, version), http.StatusSeeOther)
		return nil
	}
	recipe := struct {
		PageType     string
		LoggedInUser string
		Version      *roi.Version
	}{
		PageType:     "add",
		LoggedInUser: env.SessionUser.ID,
		Version: &roi.Version{
			Show:    show,
			Shot:    shot,
			Task:    task,
			Created: time.Now(),
		},
	}
	return executeTemplate(w, "update-version.html", recipe)
}

func updateVersionHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show", "shot", "task", "version")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	_, err = roi.GetTask(DB, show, shot, task)
	if err != nil {
		return err
	}
	version := r.FormValue("version")
	if r.Method == "POST" {
		_, err := roi.GetVersion(DB, show, shot, task, version)
		if err != nil {
			return err
		}
		timeForms, err := parseTimeForms(r.Form, "created")
		if err != nil {
			return err
		}
		mov := fmt.Sprintf("data/show/%s/%s/%s/%s/1.mov", show, shot, task, version)
		err = saveFormFile(r, "mov", mov)
		if err != nil {
			return err
		}
		u := roi.UpdateVersionParam{
			OutputFiles: fields(r.FormValue("output_files")),
			Images:      fields(r.FormValue("images")),
			WorkFile:    r.FormValue("work_file"),
			Created:     timeForms["created"],
		}
		roi.UpdateVersion(DB, show, shot, task, version, u)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return nil
	}
	v, err := roi.GetVersion(DB, show, shot, task, version)
	if err != nil {
		return err
	}
	recipe := struct {
		PageType     string
		LoggedInUser string
		Version      *roi.Version
	}{
		PageType:     "update",
		LoggedInUser: env.SessionUser.ID,
		Version:      v,
	}
	return executeTemplate(w, "update-version.html", recipe)
}
