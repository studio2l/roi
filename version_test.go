package roi

import (
	"reflect"
	"testing"
	"time"
)

var testVersionA = &Version{
	Show:     testShow.Show,
	Category: "shot",
	Group:    testGroup.Group,
	Unit:     testUnitA.Unit,
	Task:     testTaskA.Task,
	Version:  "v001",

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
		t.Fatalf("could not add project: %v", err)
	}
	defer func() {
		err = DeleteShow(db, testShow.ID())
		if err != nil {
			t.Fatalf("could not delete project: %v", err)
		}
	}()
	err = AddGroup(db, testGroup)
	if err != nil {
		t.Fatalf("could not add group to groups table: %s", err)
	}
	defer func() {
		err = DeleteGroup(db, testGroup.Show, testGroup.Category, testGroup.Group)
		if err != nil {
			t.Fatalf("could not delete group: %s", err)
		}
	}()
	err = AddUnit(db, testUnitA)
	if err != nil {
		t.Fatalf("could not add shot: %v", err)
	}
	defer func() {
		err = DeleteUnit(db, testUnitA.Show, testUnitA.Category, testUnitA.Group, testUnitA.Unit)
		if err != nil {
			t.Fatalf("could not delete shot: %v", err)
		}
	}()
	// testUnitA가 생성되면서 testTaskA도 함께 생성된다.
	defer func() {
		err = DeleteTask(db, testTaskA.Show, testTaskA.Category, testTaskA.Group, testTaskA.Unit, testTaskA.Task)
		if err != nil {
			t.Fatalf("could not delete task: %v", err)
		}
	}()
	err = UpdateTask(db, testTaskA)
	if err != nil {
		t.Fatalf("could not update task: %s", err)
	}
	err = AddVersion(db, testVersionA)
	if err != nil {
		t.Fatalf("could not add version: %v", err)
	}
	defer func() {
		err = DeleteVersion(db, testVersionA.Show, testVersionA.Category, testVersionA.Group, testVersionA.Unit, testVersionA.Task, testVersionA.Version)
		if err != nil {
			t.Fatalf("could not delete version: %v", err)
		}
	}()
	got, err := GetVersion(db, testVersionA.Show, testVersionA.Category, testVersionA.Group, testVersionA.Unit, testVersionA.Task, testVersionA.Version)
	if err != nil {
		t.Fatalf("could not get version: %v", err)
	}
	if !reflect.DeepEqual(got, testVersionA) {
		t.Fatalf("added version is not expected: got %v, want %v", got, testVersionA)
	}
	taskVersions, err := TaskVersions(db, testTaskA.Show, testTaskA.Category, testTaskA.Group, testTaskA.Unit, testTaskA.Task)
	if err != nil {
		t.Fatalf("could not get versions of task: %v", err)
	}
	if len(taskVersions) != 1 {
		t.Fatalf("task should have 1 version at this time.")
	}
	err = UpdateVersion(db, testVersionA)
	if err != nil {
		t.Fatalf("could not update version: %v", err)
	}
}
