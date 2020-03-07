package roi

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// verifyCategoryName은 받아들인 카테고리 이름이 유효하지 않다면 에러를 반환한다.
func verifyCategoryName(ctg string) error {
	switch ctg {
	case "shot", "asset":
		return nil
	}
	return fmt.Errorf("invalid category: %s", ctg)
}

type Unit struct {
	Show     string
	Category string
	Unit     string
	Status   Status
	Tags     []string
	Tasks    []string
	DueDate  time.Time
}

// verifyUnitName은 받아들인 유닛 이름이 유효하지 않다면 에러를 반환한다.
func verifyUnitName(ctg, unit string) error {
	switch ctg {
	case "shot":
		return verifyShotName(unit)
	case "asset":
		return verifyAssetName(unit)
	}
	return fmt.Errorf("invalid category: %s", ctg)
}

// SplitUnitID는 받아들인 유닛 아이디를 쇼, 카테고리, 유닛으로 분리해서 반환한다.
// 만일 유닛 아이디가 유효하지 않다면 에러를 반환한다.
func SplitUnitID(id string) (string, string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 3 {
		return "", "", "", BadRequest(fmt.Sprintf("invalid unit id: %s", id))
	}
	show := ns[0]
	ctg := ns[1]
	unit := ns[2]
	if show == "" || ctg == "" || unit == "" {
		return "", "", "", BadRequest(fmt.Sprintf("invalid unit id: %s", id))
	}
	return show, ctg, unit, nil
}

// CheckUnitExist는 해당 유닛이 존재하는지를 검사한다.
// 만일 검사중 에러가 있었다면 에러를 반환한다.
func CheckUnitExist(db *sql.DB, id string) (bool, error) {
	_, ctg, _, err := SplitUnitID(id)
	if err != nil {
		return false, err
	}
	switch ctg {
	case "shot":
		_, err := GetShot(db, id)
		if err != nil {
			return false, err
		}
		return true, nil
	case "asset":
		_, err := GetAsset(db, id)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, fmt.Errorf("invalid category: %s", ctg)
}

func UnitsHavingDue(db *sql.DB, show, ctg string) ([]*Unit, error) {
	us := make([]*Unit, 0)
	if ctg == "shot" {
		ss, err := ShotsHavingDue(db, show)
		if err != nil {
			return nil, err
		}
		for _, s := range ss {
			u := UnitFromShot(s)
			us = append(us, u)
		}
		return us, nil
	} else if ctg == "asset" {
		as, err := AssetsHavingDue(db, show)
		if err != nil {
			return nil, err
		}
		for _, a := range as {
			u := UnitFromAsset(a)
			us = append(us, u)
		}
		return us, nil
	}
	return nil, fmt.Errorf("invalid category: %s", ctg)
}
