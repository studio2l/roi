package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/studio2l/roi"
)

// showsHandler는 /shows 페이지로 사용자가 접속했을때 페이지를 반환한다.
func showsHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	shows, err := roi.AllShows(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser string
		Shows        []*roi.Show
	}{
		LoggedInUser: env.User.ID,
		Shows:        shows,
	}
	return executeTemplate(w, "shows", recipe)
}

// addShowHandler는 /add-show 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트를 생성한다.
func addShowHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return addShowPostHandler(w, r, env)
	}
	recipe := struct {
		LoggedInUser string
	}{
		LoggedInUser: env.User.ID,
	}
	return executeTemplate(w, "add-show", recipe)
}

func addShowPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	_, err = roi.GetShow(DB, id)
	if err == nil {
		return roi.BadRequest(fmt.Sprintf("show already exist: %s", id))
	} else if !errors.As(err, &roi.NotFoundError{}) {
		return err
	}
	si, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	s := &roi.Show{
		Show:              id,
		DefaultShotTasks:  si.DefaultShotTasks,
		DefaultAssetTasks: si.DefaultAssetTasks,
	}
	err = roi.AddShow(DB, s)
	if err != nil {
		return err
	}
	cfg, err := roi.GetUserConfig(DB, env.User.ID)
	if err != nil {
		return err
	}
	cfg.CurrentShow = id
	roi.UpdateUserConfig(DB, env.User.ID, cfg)
	http.Redirect(w, r, "/update-show?id="+id, http.StatusSeeOther)
	return nil
}

// updateShowHandler는 /update-show 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트 정보를 수정한다.
func updateShowHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateShowPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	p, err := roi.GetShow(DB, id)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser  string
		Show          *roi.Show
		AllShowStatus []roi.ShowStatus
	}{
		LoggedInUser:  env.User.ID,
		Show:          p,
		AllShowStatus: roi.AllShowStatus,
	}
	return executeTemplate(w, "update-show", recipe)
}

func updateShowPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	timeForms, err := parseTimeForms(r.Form,
		"vfx_due_date",
	)
	if err != nil {
		return err
	}
	s, err := roi.GetShow(DB, id)
	if err != nil {
		return err
	}
	s.Status = r.FormValue("status")
	s.VFXSupervisor = r.FormValue("vfx_supervisor")
	s.VFXManager = r.FormValue("vfx_manager")
	s.CGSupervisor = r.FormValue("cg_supervisor")
	s.VFXDueDate = timeForms["vfx_due_date"]
	s.OutputSize = r.FormValue("output_size")
	s.ViewLUT = r.FormValue("view_lut")
	s.DefaultShotTasks = fieldSplit(r.FormValue("default_shot_tasks"))
	s.DefaultAssetTasks = fieldSplit(r.FormValue("default_asset_tasks"))
	s.Tags = fieldSplit(r.FormValue("tags"))
	s.Notes = r.FormValue("notes")

	err = roi.UpdateShow(DB, id, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
