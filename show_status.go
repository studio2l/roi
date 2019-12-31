package roi

type ShowStatus string

const (
	ShowWaiting        = ShowStatus("")
	ShowPreProduction  = ShowStatus("pre")
	ShowProduction     = ShowStatus("prod")
	ShowPostProduction = ShowStatus("post")
	ShowDone           = ShowStatus("done")
	ShowHold           = ShowStatus("hold")
)

var AllShowStatus = []ShowStatus{
	ShowWaiting,
	ShowPreProduction,
	ShowProduction,
	ShowPostProduction,
	ShowDone,
	ShowHold,
}

// isValidShowStatus는 해당 태스크 상태가 유효한지를 반환한다.
func isValidShowStatus(ss ShowStatus) bool {
	for _, s := range AllShowStatus {
		if ss == s {
			return true
		}
	}
	return false
}

// UIString은 UI안에서 사용하는 현지화된 문자열이다.
// 할일: 한국어 외의 문자열 지원
func (s ShowStatus) UIString() string {
	switch s {
	case ShowWaiting:
		return "대기"
	case ShowPreProduction:
		return "프리 프로덕션"
	case ShowProduction:
		return "프로덕션"
	case ShowPostProduction:
		return "포스트 프로덕션"
	case ShowDone:
		return "완료"
	case ShowHold:
		return "홀드"
	}
	return ""
}

// UIColor는 UI안에서 사용하는 색상이다.
func (s ShowStatus) UIColor() string {
	switch s {
	case ShowWaiting:
		return ""
	case ShowPreProduction:
		return "yellow"
	case ShowProduction:
		return "yellow"
	case ShowPostProduction:
		return "green"
	case ShowDone:
		return "blue"
	case ShowHold:
		return "gray"
	}
	return ""
}
