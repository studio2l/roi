package roi

import (
	"testing"
)

var testTaskA = &Task{
	Show:     testShotA.Show,
	Shot:     testShotA.Shot,
	Task:     "fx_fire",
	Status:   TaskInProgress,
	Assignee: "kybin",
}

func TestTask(t *testing.T) {
	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddShow(db, testShow)
	if err != nil {
		t.Fatalf("could not add project: %s", err)
	}
	err = AddShot(db, testShotA)
	if err != nil {
		t.Fatalf("could not add shot: %s", err)
	}
	err = AddTask(db, testTaskA)
	if err != nil {
		t.Fatalf("could not add task: %s", err)
	}
	tasks, err := ShotTasks(db, testShotA.ID())
	if err != nil {
		t.Fatalf("could not get shot tasks: %s", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("invalid number of shot tasks: want 1, got %d", len(tasks))
	}
	tasks, err = UserTasks(db, "kybin")
	if err != nil {
		t.Fatalf("could not get user tasks: %s", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("invalid number of user tasks: want 1, got %d", len(tasks))
	}
	tasks, err = UserTasks(db, "unknown")
	if len(tasks) != 0 {
		t.Fatalf("invalid number of user tasks: want 0, got %d", len(tasks))
	}
	err = DeleteTask(db, testTaskA.ID())
	if err != nil {
		t.Fatalf("could not delete task: %s", err)
	}
	err = DeleteShot(db, testShotA.ID())
	if err != nil {
		t.Fatalf("could not delete shot: %s", err)
	}
	err = DeleteShow(db, testShow.ID())
	if err != nil {
		t.Fatalf("could not delete project: %s", err)
	}
}
