package roi

import "fmt"

type AssetStatus string

const (
	AssetWaiting    = AssetStatus("waiting")
	AssetInProgress = AssetStatus("in-progress")
	AssetDone       = AssetStatus("done")
	AssetHold       = AssetStatus("hold")
	AssetOmit       = AssetStatus("omit")
)

var AllAssetStatus = []AssetStatus{
	AssetWaiting,
	AssetInProgress,
	AssetDone,
	AssetHold,
	AssetOmit,
}

// verifyAssetStatus는 받아들인 샷의 상태가 유효하지 않다면 에러를 반환한다.
func verifyAssetStatus(ss AssetStatus) error {
	for _, s := range AllAssetStatus {
		if ss == s {
			return nil
		}
	}
	return BadRequest(fmt.Sprintf("invalid asset status: '%s'", ss))
}

func (s AssetStatus) UIString() string {
	switch s {
	case AssetWaiting:
		return "대기"
	case AssetInProgress:
		return "진행"
	case AssetDone:
		return "완료"
	case AssetHold:
		return "홀드"
	case AssetOmit:
		return "오밋"
	}
	return ""
}

func (s AssetStatus) UIColor() string {
	switch s {
	case AssetWaiting:
		return "yellow"
	case AssetInProgress:
		return "green"
	case AssetDone:
		return "blue"
	case AssetHold:
		return "grey"
	case AssetOmit:
		return "black"
	}
	return ""
}
