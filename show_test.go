package roi

import (
	"reflect"
	"testing"
	"time"
)

var testShow = &Show{
	Show:              "TEST",
	Status:            "waiting",
	Supervisor:        "김성환",
	CGSupervisor:      "김용빈",
	PD:                "조경식",
	Managers:          []string{"이한나"},
	DueDate:           time.Date(2018, 9, 31, 0, 0, 0, 0, time.Local).UTC(),
	DefaultShotTasks:  []string{},
	DefaultAssetTasks: []string{},
	Tags:              []string{},
	Notes: `테스트 쇼
클라이언트: 레이지픽처스
감독: 윤지은
PD: 김한웅
`,
	Attrs: DBStringMap{
		"output_size": "1920x1080",
		"view_lut":    "some/palce/aces.lut",
	},
}

func TestShow(t *testing.T) {
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
