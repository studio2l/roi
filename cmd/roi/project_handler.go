package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/studio2l/roi"
)

// projectsHandler는 /project 페이지로 사용자가 접속했을때 페이지를 반환한다.
func projectsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	prjs, err := roi.AllProjects(db)
	if err != nil {
		log.Print(fmt.Sprintf("error while getting projects: %s", err))
		return
	}

	session, err := getSession(r)
	if err != nil {
		log.Print(fmt.Sprintf("could not get session: %s", err))
		clearSession(w)
	}

	recipt := struct {
		LoggedInUser string
		Projects     []*roi.Project
	}{
		LoggedInUser: session["userid"],
		Projects:     prjs,
	}
	err = executeTemplate(w, "projects.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

// addProjectHandler는 /add-project 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트를 생성한다.
func addProjectHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
	if err != nil {
		http.Error(w, "could not get user information", http.StatusInternalServerError)
		clearSession(w)
		return
	}
	if u == nil {
		http.Error(w, "user not exist", http.StatusBadRequest)
		clearSession(w)
		return
	}
	if u.Role != "admin" {
		// 할일: admin이 아닌 사람은 프로젝트를 생성할 수 없도록 하기
	}
	if r.Method == "POST" {
		r.ParseForm()
		prj := r.Form.Get("project")
		if prj == "" {
			http.Error(w, "need 'project' form value", http.StatusBadRequest)
			return
		}
		exist, err := roi.ProjectExist(db, prj)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if exist {
			http.Error(w, fmt.Sprintf("project '%s' exist", prj), http.StatusBadRequest)
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
		p := &roi.Project{
			Project:       prj,
			Name:          r.Form.Get("name"),
			Status:        "waiting",
			Client:        r.Form.Get("client"),
			Director:      r.Form.Get("director"),
			Producer:      r.Form.Get("producer"),
			VFXSupervisor: r.Form.Get("vfx_supervisor"),
			VFXManager:    r.Form.Get("vfx_manager"),
			CGSupervisor:  r.Form.Get("cg_supervisor"),
			StartDate:     timeForms["start_date"],
			ReleaseDate:   timeForms["release_date"],
			CrankIn:       timeForms["crank_in"],
			CrankUp:       timeForms["crank_up"],
			VFXDueDate:    timeForms["vfx_due_date"],
			OutputSize:    r.Form.Get("output_size"),
			ViewLUT:       r.Form.Get("view_lut"),
			DefaultTasks:  fields(r.Form.Get("default_tasks"), ","),
		}
		err = roi.AddProject(db, p)
		if err != nil {
			http.Error(w, fmt.Sprintf("could not add project '%s'", p), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	recipt := struct {
		LoggedInUser string
	}{
		LoggedInUser: session["userid"],
	}
	err = executeTemplate(w, "add-project.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}

// updateProjectHandler는 /update-project 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트 정보를 수정한다.
func updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to database: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "could not get session", http.StatusUnauthorized)
		clearSession(w)
		return
	}
	u, err := roi.GetUser(db, session["userid"])
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
	r.ParseForm()
	id := r.Form.Get("id")
	if id == "" {
		http.Error(w, "need project 'id'", http.StatusBadRequest)
		return
	}
	exist, err := roi.ProjectExist(db, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if !exist {
		http.Error(w, fmt.Sprintf("project '%s' not exist", id), http.StatusBadRequest)
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
		upd := roi.UpdateProjectParam{
			Name:          r.Form.Get("name"),
			Status:        r.Form.Get("status"),
			Client:        r.Form.Get("client"),
			Director:      r.Form.Get("director"),
			Producer:      r.Form.Get("producer"),
			VFXSupervisor: r.Form.Get("vfx_supervisor"),
			VFXManager:    r.Form.Get("vfx_manager"),
			CGSupervisor:  r.Form.Get("cg_supervisor"),
			StartDate:     timeForms["start_date"],
			ReleaseDate:   timeForms["release_date"],
			CrankIn:       timeForms["crank_in"],
			CrankUp:       timeForms["crank_up"],
			VFXDueDate:    timeForms["vfx_due_date"],
			OutputSize:    r.Form.Get("output_size"),
			ViewLUT:       r.Form.Get("view_lut"),
			DefaultTasks:  fields(r.Form.Get("default_tasks"), ","),
		}
		err = roi.UpdateProject(db, id, upd)
		if err != nil {
			log.Println(err)
			http.Error(w, fmt.Sprintf("could not add project '%s'", id), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	p, err := roi.GetProject(db, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not get project: %s", id), http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.Error(w, fmt.Sprintf("could not get project: %s", id), http.StatusBadRequest)
		return
	}
	recipt := struct {
		LoggedInUser string
		Project      *roi.Project
	}{
		LoggedInUser: session["userid"],
		Project:      p,
	}
	err = executeTemplate(w, "update-project.html", recipt)
	if err != nil {
		log.Fatal(err)
	}
}
