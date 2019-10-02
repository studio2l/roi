package roi

import (
	"reflect"
	"testing"
	"time"
)

var testShow = &Show{
	Show:          "TEST",
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

func TestShow(t *testing.T) {
	want := testShow

	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddShow(db, want)
	if err != nil {
		t.Fatalf("could not add project to projects table: %s", err)
	}
	exist, err := ShowExist(db, want.Show)
	if err != nil {
		t.Fatalf("could not check project existence from projects table: %s", err)
	}
	if !exist {
		t.Fatalf("project not found from projects table: %s", want.Show)
	}
	got, err := GetShow(db, want.Show)
	if err != nil {
		t.Fatalf("could not get project from projects table: %s", err)
	}
	if !IsValidShow(got.Show) {
		if err != nil {
			t.Fatalf("find project with invalid id from projects table: %s", err)
		}
	}
	gotAll, err := AllShows(db)
	if err != nil {
		t.Fatalf("could not get all projects from projects table: %s", err)
	}
	wantAll := []*Show{want}
	if !reflect.DeepEqual(gotAll, wantAll) {
		t.Fatalf("got: %v, want: %v", got, want)
	}
	err = UpdateShow(db, want.Show, UpdateShowParam{})
	if err != nil {
		t.Fatalf("could not clear(update) project: %s", err)
	}
	err = DeleteShow(db, want.Show)
	if err != nil {
		t.Fatalf("could not delete project: %s", err)
	}
}
