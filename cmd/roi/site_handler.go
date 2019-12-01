package main

import (
	"net/http"

	"github.com/studio2l/roi"
)

func siteHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
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
		err := roi.UpdateSite(DB, s)
		if err != nil {
			return err
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return nil
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
