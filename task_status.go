package roi

import "fmt"

var AllTaskStatus = []Status{
	StatusHold,
	StatusInProgress,
	StatusDone,
}

// verifyTaskStatus는 받아들인 태스크 상태가 유효하지 않다면 에러를 반환한다.
func verifyTaskStatus(ts Status) error {
	for _, s := range AllTaskStatus {
		if ts == s {
			return nil
		}
	}
	return BadRequest(fmt.Sprintf("invalid task status: '%s'", ts))
}
