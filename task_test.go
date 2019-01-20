package roi

import (
	"database/sql"
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
	// 테스트 서버에 접속
	db, err := sql.Open("postgres", "postgresql://root@localhost:54545/roi?sslmode=disable")
	if err != nil {
		t.Fatalf("error connecting to the database: %s", err)
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
	err = DeleteTask(db, testProject.ID, testShotA.ID, testTaskA.Name)
	if err != nil {
		t.Fatalf("could not delete task: %s", err)
	}
	err = DeleteTask(db, testProject.ID, testShotA.ID, testTaskA.Name)
	if err == nil {
		t.Fatalf("could delete task again")
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
