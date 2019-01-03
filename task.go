package roi

type TaskStatus string

const (
	TaskWaiting    = TaskStatus("waiting")
	TaskAssigned   = TaskStatus("assigned")
	TaskInProgress = TaskStatus("in-progress")
	TaskPending    = TaskStatus("pending")
	TaskRetake     = TaskStatus("retake")
	TaskDone       = TaskStatus("done")
	TaskHold       = TaskStatus("hold")
	TaskOmit       = TaskStatus("omit") // 할일: task에 omit이 필요할까?
)

type Task struct {
	// 태스크 아이디. 프로젝트 내에서 고유해야 한다. ShotID.TaskName 형식.
	// 예) CG_0010.fx, EP01_SC01_0010.fx_fire
	ID string

	// 관련 아이디
	ProjectID string
	ShotID    string

	// 태스크 정보
	Name     string // 이름은 타입 또는 타입_요소로 구성된다. 예) fx, fx_fire
	Status   string
	Assignee string
}
