package roi

var AllUnitStatus = []Status{
	StatusOmit,
	StatusHold,
	StatusInProgress,
	StatusDone,
}

// verifyUnitStatus는 받아들인 샷의 상태가 유효하지 않다면 에러를 반환한다.
func verifyUnitStatus(ss Status) error {
	for _, s := range AllUnitStatus {
		if ss == s {
			return nil
		}
	}
	return BadRequest("invalid unit status: '%s'", ss)
}
