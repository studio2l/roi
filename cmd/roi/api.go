package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/studio2l/roi"
)

// apiResponse는 /api/ 하위 사이트로 사용자가 질의했을 때 json 응답을 위해 사용한다.
type apiResponse struct {
	Msg string `json:"msg"`
	Err string `json:"err"`
}

// apiOK는 api 질의가 잘 처리되었을 때
// 그 응답을 apiResponse.Msg에 담아 반환한다.
func apiOK(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusOK)
	resp, _ := json.Marshal(apiResponse{Msg: msg})
	w.Write(resp)
}

// apiInternalServerError는 내부적인 이유로 api 질의의 처리에 실패했을 때
// 이를 질의자에게 알린다. (이유는 알리지 않는다.)
func apiInternalServerError(w http.ResponseWriter) {
	resp, _ := json.Marshal(apiResponse{Err: "internal error"})
	http.Error(w, string(resp), http.StatusInternalServerError)
}

// apiBadRequest는 api 질의에 문제가 있었을 때
// 그 문제를 apiReponse.Err에 담아 반환한다.
func apiBadRequest(w http.ResponseWriter, err error) {
	if err == nil {
		// err 가 nil이어서는 안되지만, 패닉을 일으키는 것보다는 낫다.
		err = errors.New("error not explained")
	}
	resp, _ := json.Marshal(apiResponse{Err: err.Error()})
	http.Error(w, string(resp), http.StatusInternalServerError)
}

// addShotApiHander는 사용자가 api를 통해 샷을 생성할수 있도록 한다.
// 결과는 apiResponse의 json 형식으로 반환된다.
func addShotApiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	db, err := roi.DB()
	if err != nil {
		log.Printf("could not connect to db: %v", err)
		apiInternalServerError(w)
		return
	}

	prj := r.PostFormValue("project")
	if prj == "" {
		apiBadRequest(w, fmt.Errorf("'project' not specified"))
		return
	}
	exist, err := roi.ProjectExist(db, prj)
	if err != nil {
		log.Printf("could not check project %q exist: %v", prj, err)
		apiInternalServerError(w)
		return
	}
	if !exist {
		apiBadRequest(w, fmt.Errorf("project '%s' not exists", prj))
		return
	}

	id := r.PostFormValue("id")
	if id == "" {
		apiBadRequest(w, fmt.Errorf("shot 'id' not specified"))
		return
	}
	if !roi.IsValidShotID(id) {
		apiBadRequest(w, fmt.Errorf("shot id '%s' is not valid", id))
		return
	}
	exist, err = roi.ShotExist(db, prj, id)
	if err != nil {
		log.Printf("could not check shot '%s' exist: %v", id, err)
		apiInternalServerError(w)
		return
	}
	if exist {
		apiBadRequest(w, fmt.Errorf("shot '%s' already exists", id))
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
		e, err := strconv.Atoi(v)
		if err != nil {
			apiBadRequest(w, fmt.Errorf("could not convert edit_order to int: %s", v))
			return
		}
		editOrder = e
	}
	duration := 0
	v = r.PostFormValue("duration")
	if v != "" {
		d, err := strconv.Atoi(v)
		if err != nil {
			apiBadRequest(w, fmt.Errorf("could not convert duration to int: %s", v))
			return
		}
		duration = d
	}
	s := &roi.Shot{
		ID:            id,
		Status:        roi.ShotStatus(status),
		EditOrder:     editOrder,
		Description:   r.PostFormValue("description"),
		CGDescription: r.PostFormValue("cg_description"),
		TimecodeIn:    r.PostFormValue("timecode_in"),
		TimecodeOut:   r.PostFormValue("timecode_out"),
		Duration:      duration,
		Tags:          strings.Split(r.PostFormValue("tags"), ","),
	}
	err = roi.AddShot(db, prj, s)
	if err != nil {
		log.Printf("could not add shot: %v", err)
		apiInternalServerError(w)
		return
	}

	apiOK(w, fmt.Sprintf("successfully add a shot: '%s'", id))
}
