package roi

import "fmt"

var AllReviewStatus = []Status{
	StatusRetake,
	StatusApproved,
}

// verifyReviewStatus는 받아들인 리뷰 상태가 유효하지 않다면 에러를 반환한다.
func verifyReviewStatus(rs Status) error {
	for _, s := range AllReviewStatus {
		if rs == s {
			return nil
		}
	}
	return BadRequest(fmt.Sprintf("invalid review status: '%s'", rs))
}
