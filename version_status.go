package roi

import "fmt"

type VersionStatus string

const (
	VersionWaiting    = VersionStatus("waiting")
	VersionInProgress = VersionStatus("in-progress")
	VersionNeedReview = VersionStatus("need-review")
	VersionRetake     = VersionStatus("retake")
	VersionApproved   = VersionStatus("approved")
)

var AllVersionStatus = []VersionStatus{
	VersionWaiting,
	VersionInProgress,
	VersionNeedReview,
	VersionRetake,
	VersionApproved,
}

// verifyVersionStatus는 받아들인 버전 상태가 유효하지 않다면 에러를 반환한다.
func verifyVersionStatus(vs VersionStatus) error {
	for _, s := range AllVersionStatus {
		if vs == s {
			return nil
		}
	}
	return BadRequest(fmt.Sprintf("invalid version status: %s", vs))
}

// UIString은 UI안에서 사용하는 현지화된 문자열이다.
// 할일: 한국어 외의 문자열 지원
func (s VersionStatus) UIString() string {
	switch s {
	case VersionWaiting:
		return "대기중"
	case VersionInProgress:
		return "진행중"
	case VersionNeedReview:
		return "리뷰요청"
	case VersionRetake:
		return "리테이크"
	case VersionApproved:
		return "승인됨"
	}
	return ""
}
