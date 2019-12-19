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
		LoggedInUser: env.SessionUser.ID,
		Shows:        shows,
	}
	return executeTemplate(w, "shows.html", recipe)
}

// addShowHandler는 /add-show 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트를 생성한다.
func addShowHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return addShowPostHandler(w, r, env)
	}
	si, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	s := &roi.Show{
		DefaultTasks: si.DefaultTasks,
	}
	recipe := struct {
		LoggedInUser string
		Show         *roi.Show
	}{
		LoggedInUser: env.SessionUser.ID,
		Show:         s,
	}
	return executeTemplate(w, "add-show.html", recipe)
}

func addShowPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	_, err = roi.GetShow(DB, show)
	if err == nil {
		return roi.BadRequest(fmt.Sprintf("show already exist: %s", show))
	} else if !errors.As(err, &roi.NotFoundError{}) {
		return err
	}
	s := &roi.Show{
		Show: show,
	}
	err = roi.AddShow(DB, s)
	if err != nil {
		return err
	}
	cfg, err := roi.GetUserConfig(DB, env.SessionUser.ID)
	if err != nil {
		return err
	}
	cfg.CurrentShow = show
	roi.UpdateUserConfig(DB, env.SessionUser.ID, cfg)
	http.Redirect(w, r, "/update-show?show="+show, http.StatusSeeOther)
	return nil
}

// updateShowHandler는 /update-show 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트 정보를 수정한다.
func updateShowHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateShowPostHandler(w, r, env)
	}
	err := mustFields(r, "show")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	p, err := roi.GetShow(DB, show)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser  string
		Show          *roi.Show
		AllShowStatus []roi.ShowStatus
	}{
		LoggedInUser:  env.SessionUser.ID,
		Show:          p,
		AllShowStatus: roi.AllShowStatus,
	}
	return executeTemplate(w, "update-show.html", recipe)
}

func updateShowPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	timeForms, err := parseTimeForms(r.Form,
		"start_date",
		"release_date",
		"crank_in",
		"crank_up",
		"vfx_due_date",
	)
	if err != nil {
		return err
	}
	upd := roi.UpdateShowParam{
		Name:          r.FormValue("name"),
		Status:        r.FormValue("status"),
		Client:        r.FormValue("client"),
		Director:      r.FormValue("director"),
		Producer:      r.FormValue("producer"),
		VFXSupervisor: r.FormValue("vfx_supervisor"),
		VFXManager:    r.FormValue("vfx_manager"),
		CGSupervisor:  r.FormValue("cg_supervisor"),
		StartDate:     timeForms["start_date"],
		ReleaseDate:   timeForms["release_date"],
		CrankIn:       timeForms["crank_in"],
		CrankUp:       timeForms["crank_up"],
		VFXDueDate:    timeForms["vfx_due_date"],
		OutputSize:    r.FormValue("output_size"),
		ViewLUT:       r.FormValue("view_lut"),
		DefaultTasks:  fieldSplit(r.FormValue("default_tasks")),
	}
	err = roi.UpdateShow(DB, show, upd)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
