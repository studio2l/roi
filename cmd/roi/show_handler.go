package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/studio2l/roi"
)

// showsHandler는 /show 페이지로 사용자가 접속했을때 페이지를 반환한다.
func showsHandler(w http.ResponseWriter, r *http.Request) {
	shows, err := roi.AllShows(DB)
	if err != nil {
		log.Print(fmt.Sprintf("error while getting shows: %s", err))
		return
	}

	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}

	recipe := struct {
		LoggedInUser string
		Shows        []*roi.Show
	}{
		LoggedInUser: session["userid"],
		Shows:        shows,
	}
	err = executeTemplate(w, "shows.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}

// addShowHandler는 /add-show 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트를 생성한다.
func addShowHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(DB, session["userid"])
	if err != nil {
		if errors.As(err, &roi.NotFound{}) {
			handleError(w, BadRequest(err))
			return
		}
		handleError(w, Internal(err))
		return
	}
	if u.Role != "admin" {
		// 할일: admin이 아닌 사람은 프로젝트를 생성할 수 없도록 하기
	}
	if r.Method == "POST" {
		show := r.FormValue("show")
		if show == "" {
			http.Error(w, "need 'show' form value", http.StatusBadRequest)
			return
		}
		exist, err := roi.ShowExist(DB, show)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if exist {
			http.Error(w, fmt.Sprintf("show '%s' exist", show), http.StatusBadRequest)
			return
		}
		timeForms, err := parseTimeForms(r.Form,
			"start_date",
			"release_date",
			"crank_in",
			"crank_up",
			"vfx_due_date",
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		p := &roi.Show{
			Show:          show,
			Name:          r.FormValue("name"),
			Status:        "waiting",
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
			DefaultTasks:  fields(r.FormValue("default_tasks")),
		}
		err = roi.AddShow(DB, p)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not add show '%s'", p), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	si, err := roi.GetSite(DB)
	if err != nil {
		if errors.As(err, &roi.NotFound{}) {
			handleError(w, BadRequest(err))
			return
		}
		handleError(w, Internal(err))
		return
	}
	s := &roi.Show{
		DefaultTasks: si.DefaultTasks,
	}
	recipe := struct {
		LoggedInUser string
		Show         *roi.Show
	}{
		LoggedInUser: session["userid"],
		Show:         s,
	}
	err = executeTemplate(w, "add-show.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}

// updateShowHandler는 /update-show 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트 정보를 수정한다.
func updateShowHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(DB, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if false {
		// 할일: 오직 어드민, 프로젝트 슈퍼바이저, 프로젝트 매니저, CG 슈퍼바이저만
		// 이 정보를 수정할 수 있도록 하기.
		_ = u
	}
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "need show 'id'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ShowExist(DB, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("show '%s' not exist", id), http.StatusBadRequest)
		return
	}
	timeForms, err := parseTimeForms(r.Form,
		"start_date",
		"release_date",
		"crank_in",
		"crank_up",
		"vfx_due_date",
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if r.Method == "POST" {
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
			DefaultTasks:  fields(r.FormValue("default_tasks")),
		}
		err = roi.UpdateShow(DB, id, upd)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("could not add show '%s'", id), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	p, err := roi.GetShow(DB, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get show: %s", id), http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.Error(w, fmt.Sprintf("could not get show: %s", id), http.StatusBadRequest)
		return
	}
	recipe := struct {
		LoggedInUser  string
		Show          *roi.Show
		AllShowStatus []roi.ShowStatus
	}{
		LoggedInUser:  session["userid"],
		Show:          p,
		AllShowStatus: roi.AllShowStatus,
	}
	err = executeTemplate(w, "update-show.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}
