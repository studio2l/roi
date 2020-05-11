package roi

import (
	"testing"
)

var testTaskA = &Task{
	Show:     testUnitA.Show,
	Group:    testGroup.Group,
	Unit:     testUnitA.Unit,
	Task:     "fx", // testUnitA에 정의되어 있어야만 테스트가 통과한다.
	Status:   StatusInProgress,
	Assignee: "kybin",
}

func TestTask(t *testing.T) {
	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddSite(db)
	if err != nil {
		t.Fatalf("could not add site: %s", err)
	}
	defer func() {
		err := DeleteSite(db)
		if err != nil {
			t.Fatalf("could not delete site: %s", err)
		}
	}()
	err = AddShow(db, testShow)
	if err != nil {
		t.Fatalf("could not add project: %s", err)
	}
	defer func() {
		err = DeleteShow(db, testShow.ID())
		if err != nil {
			t.Fatalf("could not delete project: %s", err)
		}
	}()
	err = AddGroup(db, testGroup)
	if err != nil {
		t.Fatalf("could not add group to groups table: %s", err)
	}
	defer func() {
		err = DeleteGroup(db, testGroup.Show, testGroup.Group)
		if err != nil {
			t.Fatalf("could not delete group: %s", err)
		}
	}()
	err = AddUnit(db, testUnitA)
	if err != nil {
		t.Fatalf("could not add unit: %s", err)
	}
	defer func() {
		err = DeleteUnit(db, testUnitA.Show, testUnitA.Group, testUnitA.Unit)
		if err != nil {
			t.Fatalf("could not delete shot: %s", err)
		}
	}()
	// testUnitA가 생성되면서 testTaskA도 함께 생성된다.
	defer func() {
		err = DeleteTask(db, testTaskA.Show, testTaskA.Group, testTaskA.Unit, testTaskA.Task)
		if err != nil {
			t.Fatalf("could not delete task: %s", err)
		}
	}()
	_, err = GetTask(db, testTaskA.Show, testTaskA.Group, testTaskA.Unit, testTaskA.Task)
	if err != nil {
		t.Fatalf("could not get task: %s", testTaskA.ID())
	}
	err = UpdateTask(db, testTaskA)
	if err != nil {
		t.Fatalf("could not update task: %s", err)
	}
	tasks, err := UnitTasks(db, testUnitA.Show, testUnitA.Group, testUnitA.Unit)
	if err != nil {
		t.Fatalf("could not get shot tasks: %s", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("invalid number of unit tasks: want 1, got %d", len(tasks))
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
}
