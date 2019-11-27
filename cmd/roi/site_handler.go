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
		s := &roi.Site{
			VFXSupervisors:  fields(r.FormValue("vfx_supervisors")),
			VFXProducers:    fields(r.FormValue("vfx_producers")),
			CGSupervisors:   fields(r.FormValue("cg_supervisors")),
			ProjectManagers: fields(r.FormValue("project_managers")),
			Tasks:           fields(r.FormValue("tasks")),
			DefaultTasks:    fields(r.FormValue("default_tasks")),
			Leads:           fields(r.FormValue("leads")),
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
