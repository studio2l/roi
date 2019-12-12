package main

import (
	"net/http"

	"github.com/studio2l/roi"
)

func siteHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return sitePostHander(w, r, env)
	}
	s, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser string
		Site         *roi.Site
	}{
		LoggedInUser: env.SessionUser.ID,
		Site:         s,
	}
	return executeTemplate(w, "site.html", recipe)
}

func sitePostHander(w http.ResponseWriter, r *http.Request, env *Env) error {
	s := &roi.Site{
		VFXSupervisors:  fieldSplit(r.FormValue("vfx_supervisors")),
		VFXProducers:    fieldSplit(r.FormValue("vfx_producers")),
		CGSupervisors:   fieldSplit(r.FormValue("cg_supervisors")),
		ProjectManagers: fieldSplit(r.FormValue("project_managers")),
		Tasks:           fieldSplit(r.FormValue("tasks")),
		DefaultTasks:    fieldSplit(r.FormValue("default_tasks")),
		Leads:           fieldSplit(r.FormValue("leads")),
	}
	err := roi.UpdateSite(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
