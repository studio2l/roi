package roi

import (
	"fmt"
	"strings"
)

func verifyCategoryName(ctg string) error {
	switch ctg {
	case "shot":
		return nil
	}
	return fmt.Errorf("invalid category: %s", ctg)
}

func verifyUnitName(ctg, unit string) error {
	switch ctg {
	case "shot":
		return verifyShotName(unit)
	}
	return fmt.Errorf("invalid category: %s", ctg)
}

// SplitUnitID는 받아들인 유닛 아이디를 쇼, 카테고리, 유닛으로 분리해서 반환한다.
// 만일 유닛 아이디가 유효하지 않다면 에러를 반환한다.
func SplitUnitID(id string) (string, string, string, error) {
	ns := strings.Split(id, "/")
	if len(ns) != 3 {
		return "", "", "", BadRequest(fmt.Sprintf("invalid task id: %s", id))
	}
	show := ns[0]
	ctg := ns[1]
	unit := ns[2]
	if show == "" || ctg == "" || unit == "" {
		return "", "", "", BadRequest(fmt.Sprintf("invalid task id: %s", id))
	}
	return show, ctg, unit, nil
}
