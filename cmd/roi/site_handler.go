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
		VFXSupervisors:  formValues(r, "vfx_supervisors"),
		VFXProducers:    formValues(r, "vfx_producers"),
		CGSupervisors:   formValues(r, "cg_supervisors"),
		ProjectManagers: formValues(r, "project_managers"),
		Tasks:           formValues(r, "tasks"),
		DefaultTasks:    formValues(r, "default_tasks"),
		Leads:           formValues(r, "leads"),
	}
	err := roi.UpdateSite(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
