package roi

import (
	"database/sql"
	"log"
	"reflect"
	"testing"
	"time"
)

var testProject = &Project{
	ID:            "TEST",
	Name:          "테스트 프로젝트",
	Status:        "waiting",
	Client:        "레이지 픽처스",
	Director:      "윤지은",
	Producer:      "김한웅",
	VFXSupervisor: "김성환",
	VFXManager:    "조경식",
	CGSupervisor:  "김용빈",
	CrankIn:       time.Date(2018, 12, 31, 7, 30, 0, 0, time.Local).UTC(),
	CrankUp:       time.Date(2019, 8, 31, 19, 0, 0, 0, time.Local).UTC(),
	StartDate:     time.Date(2018, 12, 29, 0, 0, 0, 0, time.Local).UTC(),
	ReleaseDate:   time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local).UTC(),
	VFXDueDate:    time.Date(2018, 9, 31, 0, 0, 0, 0, time.Local).UTC(),
	OutputSize:    "1920x1080",
	ViewLUT:       "some/place/aces.lut",
}

func TestProject(t *testing.T) {
	want := testProject

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
	err = AddProject(db, want)
	if err != nil {
		t.Fatalf("could not add project to projects table: %s", err)
	}
	exist, err := ProjectExist(db, want.ID)
	if err != nil {
		t.Fatalf("could not check project existence from projects table: %s", err)
	}
	if !exist {
		t.Fatalf("project not found from projects table: %s", want.ID)
	}
	got, err := GetProject(db, want.ID)
	if err != nil {
		t.Fatalf("could not get project from projects table: %s", err)
	}
	if !IsValidProjectID(got.ID) {
		if err != nil {
			t.Fatalf("find project with invalid id from projects table: %s", err)
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}
	gotAll, err := SearchAllProjects(db)
	if err != nil {
		t.Fatalf("could not get all projects from projects table: %s", err)
	}
	wantAll := []*Project{want}
	if !reflect.DeepEqual(gotAll, wantAll) {
		t.Fatalf("got: %v, want: %v", got, want)
	}

	err = DeleteProject(db, want.ID)
	if err != nil {
		t.Fatalf("could not delete project: %s", err)
	}
}
