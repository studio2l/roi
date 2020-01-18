package roi

import "fmt"

type ShotStatus string

const (
	ShotWaiting    = ShotStatus("waiting")
	ShotInProgress = ShotStatus("in-progress")
	ShotDone       = ShotStatus("done")
	ShotHold       = ShotStatus("hold")
	ShotOmit       = ShotStatus("omit")
)

var AllShotStatus = []ShotStatus{
	ShotWaiting,
	ShotInProgress,
	ShotDone,
	ShotHold,
	ShotOmit,
}

// verifyShotStatus는 받아들인 샷의 상태가 유효하지 않다면 에러를 반환한다.
func verifyShotStatus(ss ShotStatus) error {
	for _, s := range AllShotStatus {
		if ss == s {
			return nil
		}
	}
	return BadRequest(fmt.Sprintf("invalid shot status: '%s'", ss))
}

func (s ShotStatus) UIString() string {
	switch s {
	case ShotWaiting:
		return "대기"
	case ShotInProgress:
		return "진행"
	case ShotDone:
		return "완료"
	case ShotHold:
		return "홀드"
	case ShotOmit:
		return "오밋"
	}
	return ""
}

func (s ShotStatus) UIColor() string {
	switch s {
	case ShotWaiting:
		return "yellow"
	case ShotInProgress:
		return "green"
	case ShotDone:
		return "blue"
	case ShotHold:
		return "grey"
	case ShotOmit:
		return "black"
	}
	return ""
}
