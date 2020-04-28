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
	us, err := roi.Users(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser string
		Site         *roi.Site
		Users        []*roi.User
	}{
		LoggedInUser: env.User.ID,
		Site:         s,
		Users:        us,
	}
	return executeTemplate(w, "site", recipe)
}

func sitePostHander(w http.ResponseWriter, r *http.Request, env *Env) error {
	s := &roi.Site{
		VFXSupervisors:    formValues(r, "vfx_supervisors"),
		VFXProducers:      formValues(r, "vfx_producers"),
		CGSupervisors:     formValues(r, "cg_supervisors"),
		ProjectManagers:   formValues(r, "project_managers"),
		ShotTasks:         formValues(r, "shot_tasks"),
		DefaultShotTasks:  formValues(r, "default_shot_tasks"),
		AssetTasks:        formValues(r, "asset_tasks"),
		DefaultAssetTasks: formValues(r, "default_asset_tasks"),
		Leads:             formValues(r, "leads"),
	}
	err := roi.UpdateSite(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
