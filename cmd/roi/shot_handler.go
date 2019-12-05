package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/studio2l/roi"
)

func addShotHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	cfg, err := roi.GetUserConfig(DB, env.SessionUser.ID)
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	if show == "" {
		// 요청이 프로젝트를 가리키지 않을 경우 사용자가
		// 보고 있던 프로젝트를 선택한다.
		show = cfg.CurrentShow
		if show == "" {
			// 사용자의 현재 프로젝트 정보가 없을때는
			// 첫번째 프로젝트를 가리킨다.
			shows, err := roi.AllShows(DB)
			if err != nil {
				return err
			}
			if len(shows) == 0 {
				return roi.BadRequest("no shows in roi")
			}
			show = shows[0].Show
		}
		http.Redirect(w, r, "/add-shot?show="+show, http.StatusSeeOther)
		return nil
	}
	sw, err := roi.GetShow(DB, show)
	if err != nil {
		return err
	}
	cfg.CurrentShow = show
	err = roi.UpdateUserConfig(DB, env.SessionUser.ID, cfg)
	if err != nil {
		return err
	}
	if r.Method == "POST" {
		err := mustFields(r, "shot")
		if err != nil {
			return err
		}
		shot := r.FormValue("shot")
		exist, err := roi.ShotExist(DB, show, shot)
		if err != nil {
			return err
		}
		if exist {
			return roi.BadRequest(fmt.Sprintf("shot exist: %s", show+"/"+shot))
		}
		tasks := fields(r.FormValue("working_tasks"))
		s := &roi.Shot{
			Shot:          shot,
			Show:          show,
			Status:        roi.ShotWaiting,
			EditOrder:     atoi(r.FormValue("edit_order")),
			Description:   r.FormValue("description"),
			CGDescription: r.FormValue("cg_description"),
			TimecodeIn:    r.FormValue("timecode_in"),
			TimecodeOut:   r.FormValue("timecode_out"),
			Duration:      atoi(r.FormValue("duration")),
			Tags:          fields(r.FormValue("tags")),
			WorkingTasks:  tasks,
		}
		err = roi.AddShot(DB, show, s)
		if err != nil {
			return err
		}
		for _, task := range tasks {
			t := &roi.Task{
				Show:    show,
				Shot:    shot,
				Task:    task,
				Status:  roi.TaskNotSet,
				DueDate: time.Time{},
			}
			roi.AddTask(DB, show, shot, t)
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return nil
	}
	recipe := struct {
		LoggedInUser string
		Show         *roi.Show
	}{
		LoggedInUser: env.SessionUser.ID,
		Show:         sw,
	}
	return executeTemplate(w, "add-shot.html", recipe)
}

func updateShotHandler(w http.ResponseWriter, r *http.Request, env *Env) error {
	err := mustFields(r, "show", "shot")
	if err != nil {
		return err
	}
	show := r.FormValue("show")
	shot := r.FormValue("shot")
	_, err = roi.GetShow(DB, show)
	if err != nil {
		return err
	}
	if r.Method == "POST" {
		_, err = roi.GetShot(DB, show, shot)
		if err != nil {
			return err
		}
		tasks := fields(r.FormValue("working_tasks"))
		tforms, err := parseTimeForms(r.Form, "due_date")
		if err != nil {
			return err
		}
		upd := roi.UpdateShotParam{
			Status:        roi.ShotStatus(r.FormValue("status")),
			EditOrder:     atoi(r.FormValue("edit_order")),
			Description:   r.FormValue("description"),
			CGDescription: r.FormValue("cg_description"),
			TimecodeIn:    r.FormValue("timecode_in"),
			TimecodeOut:   r.FormValue("timecode_out"),
			Duration:      atoi(r.FormValue("duration")),
			Tags:          fields(r.FormValue("tags")),
			WorkingTasks:  tasks,
			DueDate:       tforms["due_date"],
		}
		err = roi.UpdateShot(DB, show, shot, upd)
		if err != nil {
			return err
		}
		// 샷에 등록된 태스크 중 기존에 없었던 태스크가 있다면 생성한다.
		for _, task := range tasks {
			t := &roi.Task{
				Show:    show,
				Shot:    shot,
				Task:    task,
				Status:  roi.TaskNotSet,
				DueDate: time.Time{},
			}
			exist, err := roi.TaskExist(DB, show, shot, task)
			if err != nil {
				return err
			}
			if !exist {
				err := roi.AddTask(DB, show, shot, t)
				if err != nil {
					return err
				}
			}
		}
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return nil
	}
	s, err := roi.GetShot(DB, show, shot)
	if err != nil {
		return err
	}
	ts, err := roi.ShotTasks(DB, show, shot)
	if err != nil {
		return err
	}
	tm := make(map[string]*roi.Task)
	for _, t := range ts {
		tm[t.Task] = t
	}
	recipe := struct {
		LoggedInUser  string
		Shot          *roi.Shot
		AllShotStatus []roi.ShotStatus
		Tasks         map[string]*roi.Task
		AllTaskStatus []roi.TaskStatus
	}{
		LoggedInUser:  env.SessionUser.ID,
		Shot:          s,
		AllShotStatus: roi.AllShotStatus,
		Tasks:         tm,
		AllTaskStatus: roi.AllTaskStatus,
	}
	return executeTemplate(w, "update-shot.html", recipe)
}
