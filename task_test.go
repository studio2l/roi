package roi

import (
	"testing"
)

var testTaskA = &Task{
	Project:  testShotA.Project,
	Shot:     testShotA.Shot,
	Task:     "fx_fire",
	Status:   TaskNotSet,
	Assignee: "kybin",
}

func TestTask(t *testing.T) {
	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddProject(db, testProject)
	if err != nil {
		t.Fatalf("could not add project: %s", err)
	}
	err = AddShot(db, testTaskA.Project, testShotA)
	if err != nil {
		t.Fatalf("could not add shot: %s", err)
	}
	err = AddTask(db, testTaskA.Project, testTaskA.Shot, testTaskA)
	if err != nil {
		t.Fatalf("could not add task: %s", err)
	}
	exist, err := TaskExist(db, testTaskA.Project, testTaskA.Shot, testTaskA.Task)
	if err != nil {
		t.Fatalf("could not check task exist: %s", err)
	}
	if !exist {
		t.Fatalf("added task not exist")
	}
	tasks, err := UserTasks(db, "kybin")
	if len(tasks) != 1 {
		t.Fatalf("invalid number of user tasks: want 1, got %d", len(tasks))
	}
	tasks, err = UserTasks(db, "unknown")
	if len(tasks) != 0 {
		t.Fatalf("invalid number of user tasks: want 0, got %d", len(tasks))
	}
	err = DeleteTask(db, testTaskA.Project, testTaskA.Shot, testTaskA.Task)
	if err != nil {
		t.Fatalf("could not delete task: %s", err)
	}
	exist, err = TaskExist(db, testTaskA.Project, testTaskA.Shot, testTaskA.Task)
	if err != nil {
		t.Fatalf("could not check task exist: %s", err)
	}
	if exist {
		t.Fatalf("deleted task exist")
	}

	err = DeleteShot(db, testTaskA.Project, testTaskA.Shot)
	if err != nil {
		t.Fatalf("could not delete shot: %s", err)
	}
	err = DeleteProject(db, testTaskA.Project)
	if err != nil {
		t.Fatalf("could not delete project: %s", err)
	}
}
