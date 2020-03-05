package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/studio2l/roi"
)

// apiOK는 api 질의가 잘 처리되었을 때
// 그 응답을 roi.APIResponse.Msg에 담아 반환한다.
func apiOK(w http.ResponseWriter, msg interface{}) {
	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(roi.APIResponse{Msg: msg})
	w.Write(resp)
}

// apiInternalServerError는 내부적인 이유로 api 질의의 처리에 실패했을 때
// 이를 질의자에게 알린다. (이유는 알리지 않는다.)
func apiInternalServerError(w http.ResponseWriter) {
	resp, _ := json.Marshal(roi.APIResponse{Err: "internal error"})
	http.Error(w, string(resp), http.StatusInternalServerError)
}

// apiBadRequest는 api 질의에 문제가 있었을 때
// 그 문제를 apiReponse.Err에 담아 반환한다.
func apiBadRequest(w http.ResponseWriter, err error) {
	if err == nil {
		// err 가 nil이어서는 안되지만, 패닉을 일으키는 것보다는 낫다.
		err = errors.New("error not explained")
	}
	resp, _ := json.Marshal(roi.APIResponse{Err: err.Error()})
	http.Error(w, string(resp), http.StatusInternalServerError)
}

// addShowApiHander는 사용자가 api를 통해 프로젝트를 생성할수 있도록 한다.
// 결과는 roi.APIResponse의 json 형식으로 반환된다.
func addShowApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	show := r.PostFormValue("show")
	if show == "" {
		apiBadRequest(w, fmt.Errorf("'id' not specified"))
		return
	}
	_, err := roi.GetShow(DB, show)
	if err == nil {
		apiBadRequest(w, fmt.Errorf("show already exist: %v", show))
		return
	} else if !errors.As(err, &roi.NotFoundError{}) {
		log.Printf("could not check show %q exist: %v", show, err)
		apiInternalServerError(w)
		return
	}
	si, err := roi.GetSite(DB)
	if err != nil {
		apiInternalServerError(w)
		return
	}
	p := &roi.Show{
		Show:              show,
		DefaultShotTasks:  si.DefaultShotTasks,
		DefaultAssetTasks: si.DefaultAssetTasks,
	}
	err = roi.AddShow(DB, p)
	if err != nil {
		log.Printf("could not add show: %v", err)
		apiInternalServerError(w)
		return
	}
	apiOK(w, fmt.Sprintf("successfully add a show: '%s'", show))
}

// addShotApiHander는 사용자가 api를 통해 샷을 생성할수 있도록 한다.
// 결과는 roi.APIResponse의 json 형식으로 반환된다.
func addShotApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	id := r.PostFormValue("id")
	if id == "" {
		apiBadRequest(w, fmt.Errorf("'id' not specified"))
		return
	}
	show, shot, err := roi.SplitShotID(id)
	if err != nil {
		apiBadRequest(w, fmt.Errorf("invalid shot id: %v", id))
		return
	}
	_, err = roi.GetShot(DB, id)
	if err == nil {
		apiBadRequest(w, fmt.Errorf("shot already exist: %v", id))
		return
	} else if !errors.As(err, &roi.NotFoundError{}) {
		log.Printf("could not check shot '%s' exist: %v", id, err)
		apiInternalServerError(w)
		return
	}
	status := "waiting"
	v := r.PostFormValue("status")
	if v != "" {
		status = v
	}
	editOrder := 0
	v = r.PostFormValue("edit_order")
	if v != "" {
		// strconv.Itoa 대신 ParseFloat을 쓰는 이유는 엑셀에서 분명히 정수형으로
		// 입력한 데이터를 excelize 패키지를 통해 받을 때 실수형으로 변경되기 때문이다.
		// 무엇이 문제인지는 확실치 않지만, 실수형으로 받은후 정수로 변환하는 것이
		// 안전하다.
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			apiBadRequest(w, fmt.Errorf("could not convert duration to int: %s", v))
			return
		}
		editOrder = int(f)
	}
	duration := 0
	v = r.PostFormValue("duration")
	if v != "" {
		// 위 내용 참조
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			apiBadRequest(w, fmt.Errorf("could not convert duration to int: %s", v))
			return
		}
		duration = int(f)
	}
	tasks := fieldSplit(r.FormValue("tasks"))
	if len(tasks) == 0 {
		p, err := roi.GetShow(DB, show)
		if err != nil {
			handleError(w, err)
			return
		}
		tasks = p.DefaultShotTasks
	}
	s := &roi.Shot{
		Show:          show,
		Shot:          shot,
		Status:        roi.UnitStatus(status),
		EditOrder:     editOrder,
		Description:   r.PostFormValue("description"),
		CGDescription: r.PostFormValue("cg_description"),
		TimecodeIn:    r.PostFormValue("timecode_in"),
		TimecodeOut:   r.PostFormValue("timecode_out"),
		Duration:      duration,
		Tags:          fieldSplit(r.PostFormValue("tags")),
		Tasks:         tasks,
	}
	err = roi.AddShot(DB, s)
	if err != nil {
		log.Printf("could not add shot: %v", err)
		apiInternalServerError(w)
		return
	}
	apiOK(w, fmt.Sprintf("successfully add a shot: '%s'", shot))
}

// getShotApiHander는 사용자가 api를 통해 샷 정보를 받을수 있도록 한다.
// id 필드가 여럿 있다면 그 순서대로 샷 정보를 반환한다.
// 결과는 roi.APIResponse의 json 형식으로 반환된다.
func getShotApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := mustFields(r, "id")
	if err != nil {
		apiBadRequest(w, err)
		return
	}
	ids := r.Form["id"]
	ss := make(map[string]*roi.Shot)
	for _, id := range ids {
		s, err := roi.GetShot(DB, id)
		if err != nil {
			apiBadRequest(w, err)
			return
		}
		ss[id] = s
	}
	apiOK(w, ss)
}

// getShotTasksApiHander는 사용자가 api를 통해 샷을 생성할수 있도록 한다.
// id 필드가 여럿 있다면 그 순서대로 샷의 태스크 정보를 반환한다.
// 샷 별로 그룹을 하지는 않는다.
// 결과는 roi.APIResponse의 json 형식으로 반환된다.
func getShotTasksApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := mustFields(r, "id")
	if err != nil {
		apiBadRequest(w, err)
		return
	}
	ids := r.Form["id"]
	allTs := make(map[string][]*roi.Task)
	for _, id := range ids {
		ts, err := roi.ShotTasks(DB, id)
		if err != nil {
			apiBadRequest(w, err)
			return
		}
		allTs[id] = ts
	}
	apiOK(w, allTs)
}
