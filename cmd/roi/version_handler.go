package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/studio2l/roi"
)

func addVersionHandler(w http.ResponseWriter, r *http.Request) {
	u, err := sessionUser(r)
	if err != nil {
		if errors.As(err, &roi.NotFoundError{}) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		handleError(w, err)
		clearSession(w)
		return
	}
	err = mustFields(r, "show", "shot", "task")
	if err != nil {
		handleError(w, err)
		return
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	taskID := fmt.Sprintf("%s.%s.%s", show, shot, task)
	if r.Method == "POST" {
		err := mustFields(r, "version")
		if err != nil {
			handleError(w, err)
			return
		}
		version := r.FormValue("version")
		timeForms, err := parseTimeForms(r.Form, "created")
		if err != nil {
			handleError(w, err)
			return
		}
		_, err = roi.GetTask(DB, show, shot, task)
		if err != nil {
			handleError(w, err)
			return
		}
		mov := fmt.Sprintf("data/show/%s/%s/%s/%s/1.mov", show, shot, task, version)
		err = saveFormFile(r, "mov", mov)
		if err != nil {
			handleError(w, err)
			return
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
			handleError(w, fmt.Errorf("could not add version to task '%s': %w", taskID, err))
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/update-version?show=%s&shot=%s&task=%s&version=%s", show, shot, task, version), http.StatusSeeOther)
		return
	}
	recipe := struct {
		PageType     string
		LoggedInUser string
		Version      *roi.Version
	}{
		PageType:     "add",
		LoggedInUser: u.ID,
		Version: &roi.Version{
			Show:    show,
			Shot:    shot,
			Task:    task,
			Created: time.Now(),
		},
	}
	err = executeTemplate(w, "update-version.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func updateVersionHandler(w http.ResponseWriter, r *http.Request) {
	u, err := sessionUser(r)
	if err != nil {
		if errors.As(err, &roi.NotFoundError{}) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		handleError(w, err)
		clearSession(w)
		return
	}
	err = mustFields(r, "show", "shot", "task", "version")
	if err != nil {
		handleError(w, err)
		return
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	task := r.FormValue("task")
	_, err = roi.GetTask(DB, show, shot, task)
	if err != nil {
		handleError(w, err)
		return
	}
	version := r.FormValue("version")
	versionID := fmt.Sprintf("%s.%s.%s.%s", show, shot, task, version)
	if r.Method == "POST" {
		exist, err := roi.VersionExist(DB, show, shot, task, version)
		if err != nil {
			handleError(w, err)
			return
		}
		if !exist {
			handleError(w, fmt.Errorf("version not found: %s", versionID))
			return
		}
		timeForms, err := parseTimeForms(r.Form, "created")
		if err != nil {
			handleError(w, err)
			return
		}
		mov := fmt.Sprintf("data/show/%s/%s/%s/%s/1.mov", show, shot, task, version)
		err = saveFormFile(r, "mov", mov)
		if err != nil {
			handleError(w, err)
			return
		}
		u := roi.UpdateVersionParam{
			OutputFiles: fields(r.FormValue("output_files")),
			Images:      fields(r.FormValue("images")),
			WorkFile:    r.FormValue("work_file"),
			Created:     timeForms["created"],
		}
		roi.UpdateVersion(DB, show, shot, task, version, u)
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	v, err := roi.GetVersion(DB, show, shot, task, version)
	if err != nil {
		handleError(w, err)
		return
	}
	recipe := struct {
		PageType     string
		LoggedInUser string
		Version      *roi.Version
	}{
		PageType:     "update",
		LoggedInUser: u.ID,
		Version:      v,
	}
	err = executeTemplate(w, "update-version.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
	return
}
