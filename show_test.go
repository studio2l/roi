package roi

import (
	"reflect"
	"testing"
	"time"
)

var testShow = &Show{
	Show:              "TEST",
	Name:              "테스트 프로젝트",
	Status:            "waiting",
	Client:            "레이지 픽처스",
	Director:          "윤지은",
	Producer:          "김한웅",
	VFXSupervisor:     "김성환",
	VFXManager:        "조경식",
	CGSupervisor:      "김용빈",
	CrankIn:           time.Date(2018, 12, 31, 7, 30, 0, 0, time.Local).UTC(),
	CrankUp:           time.Date(2019, 8, 31, 19, 0, 0, 0, time.Local).UTC(),
	StartDate:         time.Date(2018, 12, 29, 0, 0, 0, 0, time.Local).UTC(),
	ReleaseDate:       time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local).UTC(),
	VFXDueDate:        time.Date(2018, 9, 31, 0, 0, 0, 0, time.Local).UTC(),
	OutputSize:        "1920x1080",
	ViewLUT:           "some/place/aces.lut",
	DefaultShotTasks:  []string{},
	DefaultAssetTasks: []string{},
}

func TestShow(t *testing.T) {
	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddShow(db, testShow)
	if err != nil {
		t.Fatalf("could not add project to projects table: %s", err)
	}
	defer func() {
		DeleteShow(db, testShow.ID())
		if err != nil {
			t.Fatalf("could not delete project: %s", err)
		}
	}()
	got, err := GetShow(db, testShow.ID())
	if err != nil {
		t.Fatalf("could not get project from projects table: %s", err)
	}
	if !reflect.DeepEqual(got, testShow) {
		t.Fatalf("got: %v, want: %v", got, testShow)
	}
	gotAll, err := AllShows(db)
	if err != nil {
		t.Fatalf("could not get all projects from projects table: %s", err)
	}
	wantAll := []*Show{testShow}
	if !reflect.DeepEqual(gotAll, wantAll) {
		t.Fatalf("got: %v, want: %v", gotAll, wantAll)
	}
	err = UpdateShow(db, testShow.ID(), testShow)
	if err != nil {
		t.Fatalf("could not update project: %s", err)
	}
}
