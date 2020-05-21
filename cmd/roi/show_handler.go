package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/studio2l/roi"
)

// showsHandler는 /shows 페이지로 사용자가 접속했을때 페이지를 반환한다.
func showsHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	shows, err := roi.AllShows(DB)
	if err != nil {
		return err
	}
	if len(shows) == 0 {
		recipe := struct {
			LoggedInUser string
		}{
			LoggedInUser: env.User.ID,
		}
		return executeTemplate(w, "no-shows", recipe)
	}
	shotGroups := make(map[string][]*roi.Group)
	assetGroups := make(map[string][]*roi.Group)
	for _, s := range shows {
		gs, err := roi.ShowGroups(DB, s.Show)
		if err != nil {
			return err
		}
		sgrps := make([]*roi.Group, 0)
		agrps := make([]*roi.Group, 0)
		for _, g := range gs {
			if strings.IndexAny(g.Group, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") == 0 {
				sgrps = append(sgrps, g)
			} else {
				agrps = append(agrps, g)
			}
		}
		shotGroups[s.Show] = sgrps
		assetGroups[s.Show] = agrps
	}
	recipe := struct {
		Env         *Env
		Shows       []*roi.Show
		ShotGroups  map[string][]*roi.Group
		AssetGroups map[string][]*roi.Group
	}{
		Env:         env,
		Shows:       shows,
		ShotGroups:  shotGroups,
		AssetGroups: assetGroups,
	}
	return executeTemplate(w, "shows", recipe)
}

// addShowHandler는 /add-show 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트를 생성한다.
func addShowHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return addShowPostHandler(w, r, env)
	}
	recipe := struct {
		Env *Env
	}{
		Env: env,
	}
	return executeTemplate(w, "add-show", recipe)
}

func addShowPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	_, err = roi.GetShow(DB, id)
	if err == nil {
		return roi.BadRequest("show already exist: %s", id)
	} else if !errors.As(err, &roi.NotFoundError{}) {
		return err
	}
	s := &roi.Show{
		Show: id,
	}
	err = roi.AddShow(DB, s)
	if err != nil {
		return err
	}
	cfg, err := roi.GetUserConfig(DB, env.User.ID)
	if err != nil {
		return err
	}
	cfg.CurrentShow = id
	roi.UpdateUserConfig(DB, env.User.ID, cfg)
	http.Redirect(w, r, "/update-show?id="+id, http.StatusSeeOther)
	return nil
}

// updateShowHandler는 /update-show 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트 정보를 수정한다.
func updateShowHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateShowPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	p, err := roi.GetShow(DB, id)
	if err != nil {
		return err
	}
	recipe := struct {
		Env           *Env
		Show          *roi.Show
		AllShowStatus []roi.ShowStatus
	}{
		Env:           env,
		Show:          p,
		AllShowStatus: roi.AllShowStatus,
	}
	return executeTemplate(w, "update-show", recipe)
}

func updateShowPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	timeForms, err := parseTimeForms(r.Form,
		"due_date",
	)
	if err != nil {
		return err
	}
	s, err := roi.GetShow(DB, id)
	if err != nil {
		return err
	}
	s.Status = r.FormValue("status")
	s.Supervisor = r.FormValue("supervisor")
	s.CGSupervisor = r.FormValue("cg_supervisor")
	s.PD = r.FormValue("pd")
	s.Managers = fieldSplit(r.FormValue("managers"))
	s.DueDate = timeForms["due_date"]
	s.Tags = fieldSplit(r.FormValue("tags"))
	s.Notes = r.FormValue("notes")
	s.Attrs = make(roi.DBStringMap)

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

	err = roi.UpdateShow(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
