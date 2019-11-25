package main

import (
	"log"
	"net/http"

	"github.com/studio2l/roi"
)

func siteHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		log.Printf("could not get session: %s", err)
		clearSession(w)
	}
	if r.Method == "POST" {
		r.ParseForm()
		s := &roi.Site{
			VFXSupervisors:  fields(r.Form.Get("vfx_supervisors")),
			VFXProducers:    fields(r.Form.Get("vfx_producers")),
			CGSupervisors:   fields(r.Form.Get("cg_supervisors")),
			ProjectManagers: fields(r.Form.Get("project_managers")),
			Tasks:           fields(r.Form.Get("tasks")),
			DefaultTasks:    fields(r.Form.Get("default_tasks")),
			Leads:           fields(r.Form.Get("leads")),
		}
		err = roi.UpdateSite(DB, s)
		if err != nil {
			log.Printf("could not update site: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}
	s, err := roi.GetSite(DB)
	if err != nil {
		log.Printf("could not get site: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	recipe := struct {
		LoggedInUser string
		Site         *roi.Site
	}{
		LoggedInUser: session["userid"],
		Site:         s,
	}
	err = executeTemplate(w, "site.html", recipe)
	if err != nil {
		log.Fatal(err)
	}
}
