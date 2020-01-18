package roi

import "fmt"

type TaskStatus string

const (
	TaskInProgress = TaskStatus("in-progress")
	TaskHold       = TaskStatus("hold")
	TaskDone       = TaskStatus("done")
)

var AllTaskStatus = []TaskStatus{
	TaskInProgress,
	TaskHold,
	TaskDone,
}

// verifyTaskStatus는 받아들인 태스크 상태가 유효하지 않다면 에러를 반환한다.
func verifyTaskStatus(ts TaskStatus) error {
	for _, s := range AllTaskStatus {
		if ts == s {
			return nil
		}
	}
	return BadRequest(fmt.Sprintf("invalid task status: '%s'", ts))
}

// UIString은 UI안에서 사용하는 현지화된 문자열이다.
// 할일: 한국어 외의 문자열 지원
func (s TaskStatus) UIString() string {
	switch s {
	case TaskInProgress:
		return "진행"
	case TaskHold:
		return "홀드"
	case TaskDone:
		return "완료"
	}
	return ""
}
