package main

import (
	"net/http"
	"strings"

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
		Env   *Env
		Site  *roi.Site
		Users []*roi.User
	}{
		Env:   env,
		Site:  s,
		Users: us,
	}
	return executeTemplate(w, "site", recipe)
}

func sitePostHander(w http.ResponseWriter, r *http.Request, env *Env) error {
	s := &roi.Site{
		VFXSupervisors:    formValues(r, "vfx_supervisors"),
		VFXProducers:      formValues(r, "vfx_producers"),
		CGSupervisors:     formValues(r, "cg_supervisors"),
		ProjectManagers:   formValues(r, "project_managers"),
		Tasks:             formValues(r, "tasks"),
		DefaultShotTasks:  formValues(r, "default_shot_tasks"),
		DefaultAssetTasks: formValues(r, "default_asset_tasks"),
		Leads:             formValues(r, "leads"),
		Notes:             r.FormValue("notes"),
		Attrs:             make(roi.DBStringMap),
	}

	for _, ln := range strings.Split(r.FormValue("attrs"), "\n") {
		kv := strings.SplitN(ln, ":", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.TrimSpace(kv[0])
		v := strings.TrimSpace(kv[1])
		if k == "" || v == "" {
			continue
		}
		s.Attrs[k] = v
	}

	err := roi.UpdateSite(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
