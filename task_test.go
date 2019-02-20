package roi

import (
	"log"
	"testing"
)

var testTaskA = &Task{
	ProjectID: testProject.ID,
	ShotID:    testShotA.ID,
	Name:      "fx_fire",
	Status:    TaskWaiting,
	Assignee:  "kybin",
}

func TestTask(t *testing.T) {
	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	if _, err := db.Exec("CREATE DATABASE IF NOT EXISTS roi"); err != nil {
		log.Fatal("error creating db 'roi': ", err)
	}
	err = InitTables(db)
	if err != nil {
		t.Fatalf("could not initialize tables: %v", err)
	}
	err = AddProject(db, testProject)
	if err != nil {
		t.Fatalf("could not add project: %s", err)
	}
	err = AddShot(db, testProject.ID, testShotA)
	if err != nil {
		t.Fatalf("could not add shot: %s", err)
	}

	err = AddTask(db, testProject.ID, testShotA.ID, testTaskA)
	if err != nil {
		t.Fatalf("could not add task: %s", err)
	}
	exist, err := TaskExist(db, testProject.ID, testShotA.ID, testTaskA.Name)
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
	err = DeleteTask(db, testProject.ID, testShotA.ID, testTaskA.Name)
	if err != nil {
		t.Fatalf("could not delete task: %s", err)
	}
	exist, err = TaskExist(db, testProject.ID, testShotA.ID, testTaskA.Name)
	if err != nil {
		t.Fatalf("could not check task exist: %s", err)
	}
	if exist {
		t.Fatalf("deleted task exist")
	}

	err = DeleteShot(db, testProject.ID, testShotA.ID)
	if err != nil {
		t.Fatalf("could not delete shot: %s", err)
	}
	err = DeleteProject(db, testProject.ID)
	if err != nil {
		t.Fatalf("could not delete project: %s", err)
	}
}
