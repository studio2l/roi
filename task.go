package roi

type TaskStatus int

const (
	TaskWaiting = TaskStatus(iota)
	TaskAssigned
	TaskInProgress
	TaskPending
	TaskRetake
	TaskDone
	TaskHold
	TaskOmit
)

type Task struct {
	Project string
	Scene   string
	Shot    string
	Type    string
	Status  string

	Assignee string
}
