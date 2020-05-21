package main

import (
	"net/http"
	"strings"
	"time"

	"github.com/studio2l/roi"
)

// unitsHandler는 /units/ 페이지로 사용자가 접속했을때 페이지를 반환한다.
func unitsHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	shows, err := roi.AllShows(DB)
	if err != nil {
		return err
	}
	if len(shows) == 0 {
		return roi.BadRequest("no shows in roi")
	}
	cfg, err := roi.GetUserConfig(DB, env.User.ID)
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
		http.Redirect(w, r, "/units?show="+show+"&q="+query, http.StatusSeeOther)
		return nil
	}
	s, err := roi.GetShow(DB, show)
	if err != nil {
		return err
	}
	cfg.CurrentShow = show
	err = roi.UpdateUserConfig(DB, env.User.ID, cfg)
	if err != nil {
		return err
	}
	if query == "?" {
		grps, err := roi.ShowGroups(DB, show)
		if err != nil {
			return err
		}
		sgrps := make([]*roi.Group, 0)
		agrps := make([]*roi.Group, 0)
		for _, g := range grps {
			if strings.IndexAny(g.Group, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") == 0 {
				sgrps = append(sgrps, g)
			} else {
				agrps = append(agrps, g)
			}
		}
		recipe := struct {
			Env         *Env
			Shows       []*roi.Show
			Query       string
			Show        string
			ShotGroups  []*roi.Group
			AssetGroups []*roi.Group
			Tags        []string
		}{
			Env:         env,
			Shows:       shows,
			Query:       query,
			Show:        s.ID(),
			ShotGroups:  sgrps,
			AssetGroups: agrps,
			Tags:        s.Tags,
		}
		return executeTemplate(w, "search-help", recipe)
	}

	grps := make([]string, 0)
	shots := make([]string, 0)
	f := make(map[string]string)
	for _, v := range strings.Fields(query) {
		kv := strings.Split(v, ":")
		if len(kv) == 1 {
			if v[len(v)-1] == '/' {
				grps = append(grps, v[:len(v)-1])
			} else {
				shots = append(shots, v)
			}
		} else {
			f[kv[0]] = kv[1]
		}
	}
	// toTime은 문자열을 time.Time으로 변경하되 에러가 나면 버린다.
	toTime := func(s string) time.Time {
		t, _ := timeFromString(s)
		return t
	}
	ss, err := roi.SearchUnits(DB, show, grps, shots, f["tag"], f["status"], f["task"], f["assignee"], f["task-status"], toTime(f["due"]))
	if err != nil {
		return err
	}
	tasks := make(map[string]map[string]*roi.Task)
	for _, s := range ss {
		ts, err := roi.UnitTasks(DB, s.Show, s.Group, s.Unit)
		if err != nil {
			return err
		}
		tm := make(map[string]*roi.Task)
		for _, t := range ts {
			tm[t.Task] = t
		}
		tasks[s.Unit] = tm
	}
	site, err := roi.GetSite(DB)
	if err != nil {
		return err
	}
	recipe := struct {
		Env           *Env
		Site          *roi.Site
		Shows         []*roi.Show
		Show          string
		Units         []*roi.Unit
		AllUnitStatus []roi.Status
		Tasks         map[string]map[string]*roi.Task
		AllTaskStatus []roi.Status
		Query         string
	}{
		Env:           env,
		Site:          site,
		Shows:         shows,
		Show:          show,
		Units:         ss,
		AllUnitStatus: roi.AllUnitStatus,
		Tasks:         tasks,
		AllTaskStatus: roi.AllTaskStatus,
		Query:         query,
	}
	return executeTemplate(w, "units", recipe)
}
