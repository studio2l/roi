package main

import (
	"fmt"
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
	ctg := r.FormValue("category")
	if ctg != "" && ctg != "shot" && ctg != "asset" {
		return fmt.Errorf("invalid category: %s", ctg)
	}
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
		http.Redirect(w, r, "/units?show="+show+"&category="+ctg+"&q="+query, http.StatusSeeOther)
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
		sgrps, err := roi.Groups(DB, show, "shot")
		if err != nil {
			return err
		}
		agrps, err := roi.Groups(DB, show, "asset")
		if err != nil {
			return err
		}
		recipe := struct {
			LoggedInUser string
			Shows        []*roi.Show
			Query        string
			Show         string
			Category     string
			ShotGroups   []*roi.Group
			AssetGroups  []*roi.Group
			Tags         []string
		}{
			LoggedInUser: env.User.ID,
			Shows:        shows,
			Query:        query,
			Show:         s.ID(),
			Category:     "",
			ShotGroups:   sgrps,
			AssetGroups:  agrps,
			Tags:         s.Tags,
		}
		return executeTemplate(w, "search-help", recipe)
	}

	grp := ""
	shots := make([]string, 0)
	f := make(map[string]string)
	for _, v := range strings.Fields(query) {
		kv := strings.Split(v, ":")
		if len(kv) == 1 {
			if v[len(v)-1] == '/' {
				grp = v[:len(v)-1]
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
	ss, err := roi.SearchUnits(DB, show, ctg, grp, shots, f["tag"], f["status"], f["task"], f["assignee"], f["task-status"], toTime(f["due"]))
	if err != nil {
		return err
	}
	tasks := make(map[string]map[string]*roi.Task)
	for _, s := range ss {
		ts, err := roi.UnitTasks(DB, s.Show, s.Category, s.Group, s.Unit)
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
		LoggedInUser  string
		Site          *roi.Site
		Category      string
		Shows         []*roi.Show
		Show          string
		Units         []*roi.Unit
		AllUnitStatus []roi.Status
		Tasks         map[string]map[string]*roi.Task
		AllTaskStatus []roi.Status
		Query         string
	}{
		LoggedInUser:  env.User.ID,
		Site:          site,
		Category:      ctg,
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
