package roi

import (
	"reflect"
	"testing"
	"time"
)

var testVersionA = &Version{
	Show:    testShow.Show,
	Shot:    testShotA.Shot,
	Task:    testTaskA.Task,
	Version: "v001",

	Status:      VersionInProgress,
	Owner:       "admin",
	OutputFiles: []string{"/project/test/FOO_0010/scene/test.v001.abc"},
	Images: []string{
		"/project/test/FOO_0010/render/test.v001.0001.jpg",
		"/project/test/FOO_0010/render/test.v001.0002.jpg",
	},
	Mov:       "/project/test/FOO_0010/render/test.v001.mov",
	WorkFile:  "/project/test/FOO_0010/scene/test.v001.hip",
	StartDate: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
	EndDate:   time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
}

func TestVersion(t *testing.T) {
	// 테스트 서버에 접속
	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddShow(db, testShow)
	if err != nil {
		t.Fatalf("could not add project: %v", err)
	}
	err = AddShot(db, testShotA)
	if err != nil {
		t.Fatalf("could not add shot: %v", err)
	}
	err = AddTask(db, testTaskA)
	if err != nil {
		t.Fatalf("could not add task: %v", err)
	}
	err = AddVersion(db, testVersionA)
	if err != nil {
		t.Fatalf("could not add version: %v", err)
	}
	got, err := GetVersion(db, testVersionA.ID())
	if err != nil {
		t.Fatalf("could not get version: %v", err)
	}
	if !reflect.DeepEqual(got, testVersionA) {
		t.Fatalf("added version is not expected: got %v, want %v", got, testVersionA)
	}
	shotVersions, err := ShotVersions(db, testShotA.ID())
	if err != nil {
		t.Fatalf("could not get versions of shot: %v", err)
	}
	if len(shotVersions) != 1 {
		t.Fatalf("shot should have 1 version at this time.")
	}
	taskVersions, err := TaskVersions(db, testTaskA.ID())
	if err != nil {
		t.Fatalf("could not get versions of task: %v", err)
	}
	if len(taskVersions) != 1 {
		t.Fatalf("task should have 1 version at this time.")
	}
	err = UpdateVersion(db, testVersionA.ID(), testVersionA)
	if err != nil {
		t.Fatalf("could not update version: %v", err)
	}
	err = DeleteVersion(db, testVersionA.ID())
	if err != nil {
		t.Fatalf("could not delete version: %v", err)
	}
	err = DeleteTask(db, testTaskA.ID())
	if err != nil {
		t.Fatalf("could not delete task: %v", err)
	}
	err = DeleteShot(db, testShotA.ID())
	if err != nil {
		t.Fatalf("could not delete shot: %v", err)
	}
	err = DeleteShow(db, testShow.ID())
	if err != nil {
		t.Fatalf("could not delete project: %v", err)
	}
}
