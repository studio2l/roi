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
	Project string
	Scene   string
	Shot    string
	Type    string
	Status  string

	Assignee string
}
