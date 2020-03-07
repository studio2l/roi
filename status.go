package roi

// Status는 유닛 및 태스크의 상태이다.
type Status string

// 각 항목들은 여기에 선언되어있는 상태중 일부만 사용한다.
const (
	StatusOmit       = Status("omit")
	StatusHold       = Status("hold")
	StatusInProgress = Status("in-progress")
	StatusNeedReview = Status("need-review")
	StatusRetake     = Status("retake")
	StatusApproved   = Status("approved")
	StatusDone       = Status("done")
)

var AllStatus = []Status{
	StatusOmit,
	StatusHold,
	StatusInProgress,
	StatusNeedReview,
	StatusRetake,
	StatusApproved,
	StatusDone,
}

// UIStirng은 UI에서 각 상태를 뜻할 문자열을 의미한다.
func (s Status) UIString() string {
	switch s {
	case StatusOmit:
		return "오밋"
	case StatusHold:
		return "홀드"
	case StatusInProgress:
		return "진행"
	case StatusNeedReview:
		return "리뷰대기"
	case StatusRetake:
		return "리테이크"
	case StatusApproved:
		return "승인"
	case StatusDone:
		return "완료"
	}
	return ""
}

// UIColor는 UI에서 각 상태를 뜻할 색상을 의미한다.
func (s Status) UIColor() string {
	switch s {
	case StatusOmit:
		return "black"
	case StatusHold:
		return "grey"
	case StatusInProgress:
		return "green"
	case StatusNeedReview:
		return "magenta"
	case StatusRetake:
		return "crimson"
	case StatusApproved:
		return "aquamarine"
	case StatusDone:
		return "blue"
	}
	return ""
}
