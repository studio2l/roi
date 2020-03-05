package roi

import "fmt"

type UnitStatus string

const (
	UnitOmit       = UnitStatus("omit")
	UnitHold       = UnitStatus("hold")
	UnitInProgress = UnitStatus("in-progress")
	UnitDone       = UnitStatus("done")
)

var AllUnitStatus = []UnitStatus{
	UnitOmit,
	UnitHold,
	UnitInProgress,
	UnitDone,
}

// verifyUnitStatus는 받아들인 유닛의 상태가 유효하지 않다면 에러를 반환한다.
func verifyUnitStatus(ss UnitStatus) error {
	for _, s := range AllUnitStatus {
		if ss == s {
			return nil
		}
	}
	return BadRequest(fmt.Sprintf("invalid shot status: '%s'", ss))
}

func (s UnitStatus) UIString() string {
	switch s {
	case UnitHold:
		return "홀드"
	case UnitInProgress:
		return "진행"
	case UnitOmit:
		return "오밋"
	case UnitDone:
		return "완료"
	}
	return ""
}

func (s UnitStatus) UIColor() string {
	switch s {
	case UnitHold:
		return "grey"
	case UnitInProgress:
		return "green"
	case UnitOmit:
		return "black"
	case UnitDone:
		return "blue"
	}
	return ""
}
