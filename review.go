package roi

import (
	"time"
)

type Review struct {
	// ID는 리뷰의 아이디이다. 프로젝트 내에서 고유해야 한다.
	// Shot.ID + "." + Task.Name + ".v" + pads(Output.Version, 3) + ".r" + itoa(Num)
	// 예) CG_0010.fx.v001.r1
	ID string

	ProjectID string
	OutputID  string
	UserID    string

	Num      int       // 리뷰 번호
	Reviewer User      // 리뷰 한 사람
	Msg      string    // 리뷰 내용. 텍스트거나 HTML일 수도 있다.
	Time     time.Time // 생성, 수정된 시간
}
