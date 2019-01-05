package roi

import "time"

// Output은 특정 태스크의 해당 버전 작업 결과물이다.
type Output struct {
	// ID는 작업 결과물의 아이디이다. 프로젝트 내에서 고유해야 한다.
	// Shot.ID + "." + Task.Name + ".v" + pads(Version, 3)
	// 예) CG_0010.fx.v001
	ID string

	ProjectID string
	TaskID    string

	Version  int       // 결과물 버전
	File     string    // 결과물 경로
	Mov      string    // 결과물을 영상으로 볼 수 있는 경로
	WorkFile string    // 이 결과물을 만든 작업 파일
	Time     time.Time // 결과물이 만들어진 시간
}
