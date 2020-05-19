package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/studio2l/roi"
)

// addGroupHandler는 /add-group 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트를 생성한다.
func addGroupHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return addGroupPostHandler(w, r, env)
	}
	w.Header().Set("Cache-control", "no-cache")
	cfg, err := roi.GetUserConfig(DB, env.User.ID)
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	if show == "" {
		show = cfg.CurrentShow
		if show == "" {
			// 사용자의 현재 프로젝트 정보가 없을때는
			// 첫번째 프로젝트를 가리킨다.
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
			show = shows[0].Show
		}
	}
	cfg.CurrentShow = show
	err = roi.UpdateUserConfig(DB, env.User.ID, cfg)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser string
		Show         string
	}{
		LoggedInUser: env.User.ID,
		Show:         show,
	}
	return executeTemplate(w, "add-group", recipe)
}

func addGroupPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show", "group")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	grp := r.FormValue("group")
	_, err = roi.GetGroup(DB, show, grp)
	if err == nil {
		return roi.BadRequest("group already exist: %s", roi.JoinGroupID(show, grp))
	} else if !errors.As(err, &roi.NotFoundError{}) {
		return err
	}
	si, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	defaultTasks := si.DefaultShotTasks
	if strings.IndexAny(grp, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != 0 {
		// 대문자로 시작하지 않는 그룹은 애셋 그룹이다.
		defaultTasks = si.DefaultAssetTasks
	}
	s := &roi.Group{
		Show:         show,
		Group:        grp,
		DefaultTasks: defaultTasks,
	}
	err = roi.AddGroup(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, "/update-group?id="+roi.JoinGroupID(show, grp), http.StatusSeeOther)
	return nil
}

// updateGroupHandler는 /update-group 페이지로 사용자가 접속했을때 페이지를 반환한다.
// 만일 POST로 프로젝트 정보가 오면 프로젝트 정보를 수정한다.
func updateGroupHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method == "POST" {
		return updateGroupPostHandler(w, r, env)
	}
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, err := roi.SplitGroupID(id)
	if err != nil {
		return err
	}
	p, err := roi.GetGroup(DB, show, grp)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser string
		Group        *roi.Group
	}{
		LoggedInUser: env.User.ID,
		Group:        p,
	}
	return executeTemplate(w, "update-group", recipe)
}

func updateGroupPostHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "id")
	if err != nil {
		return err
	}
	id := r.FormValue("id")
	show, grp, err := roi.SplitGroupID(id)
	if err != nil {
		return err
	}
	s, err := roi.GetGroup(DB, show, grp)
	if err != nil {
		return err
	}
	s.DefaultTasks = fieldSplit(r.FormValue("default_tasks"))
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

	err = roi.UpdateGroup(DB, s)
	if err != nil {
		return err
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
