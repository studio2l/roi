package roi

import (
	"reflect"
	"testing"
	"time"
)

var testVersionA = &Version{
	Project: testProject.Project,
	Shot:    testShotA.Shot,
	Task:    testTaskA.Task,

	Version:     0, // DB에 Version을 추가할 때는 Version이 지정되어 있으면 안된다.
	OutputFiles: []string{"/project/test/FOO_0010/scene/test.v001.abc"},
	Images: []string{
		"/project/test/FOO_0010/render/test.v001.0001.jpg",
		"/project/test/FOO_0010/render/test.v001.0002.jpg",
	},
	Mov:      "/project/test/FOO_0010/render/test.v001.mov",
	WorkFile: "/project/test/FOO_0010/scene/test.v001.hip",
	Created:  time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
}

func TestVersion(t *testing.T) {
	// 테스트 서버에 접속
	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddProject(db, testProject)
	if err != nil {
		t.Fatalf("could not add project: %v", err)
	}
	err = AddShot(db, testVersionA.Project, testShotA)
	if err != nil {
		t.Fatalf("could not add shot: %v", err)
	}
	err = AddTask(db, testVersionA.Project, testVersionA.Shot, testTaskA)
	if err != nil {
		t.Fatalf("could not add task: %v", err)
	}

	err = AddVersion(db, testVersionA.Project, testVersionA.Shot, testVersionA.Task, testVersionA)
	if err != nil {
		t.Fatalf("could not add version: %v", err)
	}
	exist, err := VersionExist(db, testVersionA.Project, testVersionA.Shot, testVersionA.Task, testVersionA.Version)
	if err != nil {
		t.Fatalf("could not check version exist: %v", err)
	}
	if !exist {
		t.Fatalf("added version not exist")
	}
	want := testVersionA
	want.Version = 1 // DB에 들어가면서 버전 번호가 추가되어야 한다.
	got, err := GetVersion(db, testVersionA.Project, testVersionA.Shot, testVersionA.Task, want.Version)
	if err != nil {
		t.Fatalf("could not get version: %v", err)
	}
	shotVersions, err := ShotVersions(db, testVersionA.Project, testVersionA.Shot)
	if err != nil {
		t.Fatalf("could not get versions of shot: %v", err)
	}
	if len(shotVersions) != 1 {
		t.Fatalf("shot should have 1 version at this time.")
	}
	taskVersions, err := TaskVersions(db, testVersionA.Project, testVersionA.Shot, testVersionA.Task)
	if err != nil {
		t.Fatalf("could not get versions of task: %v", err)
	}
	if len(taskVersions) != 1 {
		t.Fatalf("task should have 1 version at this time.")
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("added version is not expected: got %v, want %v", got, want)
	}
	err = UpdateVersion(db, testVersionA.Project, testVersionA.Shot, testVersionA.Task, testVersionA.Version, UpdateVersionParam{})
	if err != nil {
		t.Fatalf("could not clear(update) version: %v", err)
	}
	err = DeleteVersion(db, testVersionA.Project, testVersionA.Shot, testVersionA.Task, testVersionA.Version)
	if err != nil {
		t.Fatalf("could not delete version: %v", err)
	}
	exist, err = VersionExist(db, testVersionA.Project, testVersionA.Shot, testVersionA.Task, testVersionA.Version)
	if err != nil {
		t.Fatalf("could not check version exist: %v", err)
	}
	if exist {
		t.Fatalf("deleted version exist")
	}

	err = DeleteTask(db, testVersionA.Project, testVersionA.Shot, testVersionA.Task)
	if err != nil {
		t.Fatalf("could not delete task: %v", err)
	}
	err = DeleteShot(db, testVersionA.Project, testVersionA.Shot)
	if err != nil {
		t.Fatalf("could not delete shot: %v", err)
	}
	err = DeleteProject(db, testVersionA.Project)
	if err != nil {
		t.Fatalf("could not delete project: %v", err)
	}
}
