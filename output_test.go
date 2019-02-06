package roi

import (
	"database/sql"
	"log"
	"reflect"
	"testing"
	"time"
)

var testOutputA = &Output{
	ProjectID: testProject.ID,
	ShotID:    testShotA.ID,
	TaskName:  testTaskA.Name,

	Version: 0, // DB에 Output을 추가할 때는 버전이 지정되어 있으면 안된다.
	Files:   []string{"/project/test/FOO_0010/scene/test.v001.abc"},
	Images: []string{
		"/project/test/FOO_0010/render/test.v001.0001.jpg",
		"/project/test/FOO_0010/render/test.v001.0002.jpg",
	},
	Mov:      "/project/test/FOO_0010/render/test.v001.mov",
	WorkFile: "/project/test/FOO_0010/scene/test.v001.hip",
	Time:     time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
}

func TestOutput(t *testing.T) {
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
		t.Fatalf("could not add project: %v", err)
	}
	err = AddShot(db, testProject.ID, testShotA)
	if err != nil {
		t.Fatalf("could not add shot: %v", err)
	}
	err = AddTask(db, testProject.ID, testShotA.ID, testTaskA)
	if err != nil {
		t.Fatalf("could not add task: %v", err)
	}

	err = AddOutput(db, testProject.ID, testShotA.ID, testTaskA.Name, testOutputA)
	if err != nil {
		t.Fatalf("could not add output: %v", err)
	}
	exist, err := OutputExist(db, testProject.ID, testShotA.ID, testTaskA.Name, testOutputA.Version)
	if err != nil {
		t.Fatalf("could not check output exist: %v", err)
	}
	if !exist {
		t.Fatalf("added output not exist")
	}
	want := testOutputA
	want.Version = 1 // DB에 들어가면서 버전이 추가되어야 한다.
	got, err := GetOutput(db, testProject.ID, testShotA.ID, testTaskA.Name, want.Version)
	if err != nil {
		t.Fatalf("could not get output: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("added output is not expected: got %v, want %v", got, want)
	}
	err = DeleteOutput(db, testProject.ID, testShotA.ID, testTaskA.Name, testOutputA.Version)
	if err != nil {
		t.Fatalf("could not delete output: %v", err)
	}
	err = DeleteOutput(db, testProject.ID, testShotA.ID, testTaskA.Name, testOutputA.Version)
	if err == nil {
		t.Fatalf("could delete output again")
	}
	exist, err = OutputExist(db, testProject.ID, testShotA.ID, testTaskA.Name, testOutputA.Version)
	if err != nil {
		t.Fatalf("could not check output exist: %v", err)
	}
	if exist {
		t.Fatalf("deleted output exist")
	}

	err = DeleteTask(db, testProject.ID, testShotA.ID, testTaskA.Name)
	if err != nil {
		t.Fatalf("could not delete task: %v", err)
	}
	err = DeleteShot(db, testProject.ID, testShotA.ID)
	if err != nil {
		t.Fatalf("could not delete shot: %v", err)
	}
	err = DeleteProject(db, testProject.ID)
	if err != nil {
		t.Fatalf("could not delete project: %v", err)
	}
}
