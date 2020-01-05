package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// shotsHandler는 /shots/ 페이지로 사용자가 접속했을때 페이지를 반환한다.
func shotsHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	shows, err := roi.AllShows(DB)
	if err != nil {
		return err
	}
	if len(shows) == 0 {
		return roi.BadRequest("no shows in roi")
	}
	cfg, err := roi.GetUserConfig(DB, env.SessionUser.ID)
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	query := r.FormValue("q")
	if show == "" {
		// 요청이 프로젝트를 가리키지 않을 경우 사용자가
		// 보고 있던 프로젝트를 선택한다.
		show = cfg.CurrentShow
		if show == "" {
			// 사용자의 현재 프로젝트 정보가 없을때는
			// 첫번째 프로젝트를 가리킨다.
			show = shows[0].Show
		}
		http.Redirect(w, r, "/shots?show="+show+"&q="+query, http.StatusSeeOther)
		return nil
	}
	_, err = roi.GetShow(DB, show)
	if err != nil {
		return err
	}
	cfg.CurrentShow = show
	err = roi.UpdateUserConfig(DB, env.SessionUser.ID, cfg)
	if err != nil {
		return err
	}

	f := make(map[string]string)
	for _, v := range strings.Fields(query) {
		kv := strings.Split(v, ":")
		if len(kv) == 1 {
			f["shot"] = v
		} else {
			f[kv[0]] = kv[1]
		}
	}
	// toTime은 문자열을 time.Time으로 변경하되 에러가 나면 버린다.
	toTime := func(s string) time.Time {
		t, _ := timeFromString(s)
		return t
	}
	shots, err := roi.SearchShots(DB, show, f["shot"], f["tag"], f["status"], f["task"], f["assignee"], f["task-status"], toTime(f["task-due"]))
	if err != nil {
		return err
	}
	tasks := make(map[string]map[string]*roi.Task)
	for _, s := range shots {
		ts, err := roi.ShotTasks(DB, s.ID())
		if err != nil {
			return err
		}
		tm := make(map[string]*roi.Task)
		for _, t := range ts {
			tm[t.Task] = t
		}
		tasks[s.Shot] = tm
	}
	site, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		LoggedInUser  string
		Site          *roi.Site
		Shows         []*roi.Show
		Show          string
		Shots         []*roi.Shot
		AllShotStatus []roi.ShotStatus
		Tasks         map[string]map[string]*roi.Task
		AllTaskStatus []roi.TaskStatus
		Query         string
	}{
		LoggedInUser:  env.SessionUser.ID,
		Site:          site,
		Shows:         shows,
		Show:          show,
		Shots:         shots,
		AllShotStatus: roi.AllShotStatus,
		Tasks:         tasks,
		AllTaskStatus: roi.AllTaskStatus,
		Query:         query,
	}
	return executeTemplate(w, "shots.html", recipe)
}

func updateMultiShotsHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	if r.Method != "POST" {
		return roi.BadRequest("only post method allowed")
	}
	err := mustFields(r, "ids", "key", "val")
	if err != nil {
		return err
	}
	ids := strings.FieldsFunc(r.FormValue("ids"), func(c rune) bool { return (c == ' ') || (c == ',') })
	key := r.FormValue("key")
	val := r.FormValue("val")
	switch key {
	case "tag":
		for _, id := range ids {
			s, err := roi.GetShot(DB, id)
			if err != nil {
				return err
			}
			typ := r.FormValue("typ")
			if typ == "add" {
				find := false
				for _, t := range s.Tags {
					if t == val {
						find = true
					}
				}
				if !find {
					s.Tags = append(s.Tags, val)
				}
			} else if typ == "sub" {
				tags := s.Tags
				s.Tags = []string{}
				for _, t := range tags {
					if t != val {
						s.Tags = append(s.Tags, t)
					}
				}
			} else {
				return roi.BadRequest(fmt.Sprintf("%s is not a valid typ", typ))
			}
			err = roi.UpdateShotAll(DB, s)
			if err != nil {
				return err
			}
		}
	default:
		return roi.BadRequest(fmt.Sprintf("%s is not a valid key", key))
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	return nil
}
