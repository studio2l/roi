package main

import (
	"fmt"
	"net/http"

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
	show, grp, unit, task, err := roi.SplitTaskID(id)
	if err != nil {
		return err
	}
	recipe := struct {
		PageType string // 할일: 안쓰고 있으니 지울것.
		Env      *Env
		Version  *roi.Version
	}{
		Env: env,
		Version: &roi.Version{
			Show:  show,
			Group: grp,
			Unit:  unit,
			Task:  task,
		},
	}
	return executeTemplate(w, "add-version", recipe)
}

func addVersionPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id", "version")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, unit, task, err := roi.SplitTaskID(id)
	if err != nil {
		return err
	}
	version := r.FormValue("version")
	v := &roi.Version{
		Show:    show,
		Group:   grp,
		Unit:    unit,
		Task:    task,
		Version: version,
		Owner:   env.User.ID,
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
	show, grp, unit, task, ver, err := roi.SplitVersionID(id)
	if err != nil {
		return err
	}
	v, err := roi.GetVersion(DB, show, grp, unit, task, ver)
	if err != nil {
		return err
	}
	t, err := roi.GetTask(DB, show, grp, unit, task)
	if err != nil {
		return err
	}
	recipe := struct {
		Env              *Env
		Version          *roi.Version
		IsWorkingVersion bool
		IsPublishVersion bool
	}{
		Env:              env,
		Version:          v,
		IsWorkingVersion: t.WorkingVersion == v.Version,
		IsPublishVersion: t.PublishVersion == v.Version,
	}
	return executeTemplate(w, "update-version", recipe)
}

func updateVersionPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, unit, task, ver, err := roi.SplitVersionID(id)
	if err != nil {
		return err
	}
	dstd := fmt.Sprintf("data/show/%s", id)
	err = saveFormFiles(r, "preview_files", dstd)
	if err != nil {
		return err
	}
	v, err := roi.GetVersion(DB, show, grp, unit, task, ver)
	if err != nil {
		return err
	}
	v.OutputFiles = fieldSplit(r.FormValue("output_files"))
	v.Images = fieldSplit(r.FormValue("images"))
	v.WorkFile = r.FormValue("work_file")

	err = roi.UpdateVersion(DB, v)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
