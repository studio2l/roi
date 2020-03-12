package roi

import (
	"database/sql"
	"fmt"
	"time"
)

// ReviewTarget은 리뷰의 대상이 되는 유닛 또는 태스크 정보이다.
type ReviewTarget struct {
	Kind    string // "shot", "asset" 또는 "task"
	Show    string
	Name    string // 만일 Kind가 유닛이라면 유닛명, 태스크라면 유닛명 + "/" + 태스크명이 이름이다.
	Status  Status
	DueDate time.Time
}

// ReviewTargetsHavingDue는 마감일을 가진 ReviewTarget을 반환한다.
func ReviewTargetsHavingDue(db *sql.DB, show, ctg, kind string) ([]*ReviewTarget, error) {
	rts := make([]*ReviewTarget, 0)

	if kind != "unit" && kind != "task" {
		return nil, BadRequest(fmt.Sprintf("invalid kind: %s", kind))
	}
	if kind == "task" {
		ts, err := TasksHavingDue(db, show, ctg)
		if err != nil {
			return nil, err
		}
		for _, t := range ts {
			rts = append(rts, ReviewTargetFromTask(t))
		}
		return rts, nil
	}

	// 이제 kind는 unit이다.
	switch ctg {
	case "shot":
		ss, err := ShotsHavingDue(db, show)
		if err != nil {
			return nil, err
		}
		for _, s := range ss {
			rt := ReviewTargetFromShot(s)
			rts = append(rts, rt)
		}
	case "asset":
		as, err := AssetsHavingDue(db, show)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			rt := ReviewTargetFromAsset(a)
			rts = append(rts, rt)
		}
	default:
		return nil, BadRequest(fmt.Sprintf("invalid category: %s", ctg))
	}
	return rts, nil
}

// ReviewTargetsNeedReview는 마감일을 가진 ReviewTarget을 반환한다.
func ReviewTargetsNeedReview(db *sql.DB, show, ctg, kind string) ([]*ReviewTarget, error) {
	rts := make([]*ReviewTarget, 0)

	if kind != "unit" && kind != "task" {
		return nil, BadRequest(fmt.Sprintf("invalid kind: %s", kind))
	}
	if kind == "task" {
		ts, err := TasksNeedReview(db, show, ctg)
		if err != nil {
			return nil, err
		}
		for _, t := range ts {
			rts = append(rts, ReviewTargetFromTask(t))
		}
		return rts, nil
	}

	// 이제 kind는 unit이다.
	switch ctg {
	case "shot":
		ss, err := ShotsNeedReview(db, show)
		if err != nil {
			return nil, err
		}
		for _, s := range ss {
			rt := ReviewTargetFromShot(s)
			rts = append(rts, rt)
		}
	case "asset":
		as, err := AssetsNeedReview(db, show)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			rt := ReviewTargetFromAsset(a)
			rts = append(rts, rt)
		}
	default:
		return nil, BadRequest(fmt.Sprintf("invalid category: %s", ctg))
	}
	return rts, nil
}

type Review struct {
	// ID는 리뷰의 아이디이다. 프로젝트 내에서 고유해야 한다.
	// Shot.ID + "." + Task.Name + ".v" + pads(Output.Version, 3) + ".r" + itoa(Num)
	// 예) CG_0010.fx.v001.r1
	ID string

	ProjectID string
	OutputID  string
	UserID    string

	Num      int       // 리뷰 번호
	Reviewer User      // 리뷰 한 사람
	Msg      string    // 리뷰 내용. 텍스트거나 HTML일 수도 있다.
	Time     time.Time // 생성, 수정된 시간
}
