package roi

import (
	"reflect"
	"testing"
	"time"
)

var testVersionA = &Version{
	ProjectID: testProject.ID,
	ShotID:    testShotA.ID,
	TaskName:  testTaskA.Name,

	Num:         0, // DB에 Version을 추가할 때는 Num이 지정되어 있으면 안된다.
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
	err = AddShot(db, testProject.ID, testShotA)
	if err != nil {
		t.Fatalf("could not add shot: %v", err)
	}
	err = AddTask(db, testProject.ID, testShotA.ID, testTaskA)
	if err != nil {
		t.Fatalf("could not add task: %v", err)
	}

	err = AddVersion(db, testProject.ID, testShotA.ID, testTaskA.Name, testVersionA)
	if err != nil {
		t.Fatalf("could not add version: %v", err)
	}
	exist, err := VersionExist(db, testProject.ID, testShotA.ID, testTaskA.Name, testVersionA.Num)
	if err != nil {
		t.Fatalf("could not check version exist: %v", err)
	}
	if !exist {
		t.Fatalf("added version not exist")
	}
	want := testVersionA
	want.Num = 1 // DB에 들어가면서 버전 번호가 추가되어야 한다.
	got, err := GetVersion(db, testProject.ID, testShotA.ID, testTaskA.Name, want.Num)
	if err != nil {
		t.Fatalf("could not get version: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("added version is not expected: got %v, want %v", got, want)
	}
	err = UpdateVersion(db, testProject.ID, testShotA.ID, testTaskA.Name, testVersionA.Num, UpdateVersionParam{})
	if err != nil {
		t.Fatalf("could not clear(update) version: %v", err)
	}
	err = DeleteVersion(db, testProject.ID, testShotA.ID, testTaskA.Name, testVersionA.Num)
	if err != nil {
		t.Fatalf("could not delete version: %v", err)
	}
	exist, err = VersionExist(db, testProject.ID, testShotA.ID, testTaskA.Name, testVersionA.Num)
	if err != nil {
		t.Fatalf("could not check version exist: %v", err)
	}
	if exist {
		t.Fatalf("deleted version exist")
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
