package roi

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

// isValidShotStatus는 해당 샷 상태가 유효한지를 반환한다.
func isValidShotStatus(ss ShotStatus) bool {
	for _, s := range AllShotStatus {
		if ss == s {
			return true
		}
	}
	return false
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
