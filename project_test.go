package roi

import (
	"reflect"
	"testing"
	"time"
)

func TestOrdMapFromProject(t *testing.T) {
	p := Project{
		ID:            "TEST",
		Name:          "테스트 프로젝트",
		Status:        "waiting",
		Client:        "레이지 픽처스",
		Director:      "윤지은",
		Producer:      "김한웅",
		VFXSupervisor: "김성환",
		VFXManager:    "조경식",
		CGSupervisor:  "김용빈",
		CrankIn:       time.Date(2018, 12, 31, 7, 30, 0, 0, time.Local),
		CrankUp:       time.Date(2019, 8, 31, 19, 0, 0, 0, time.Local),
		StartDate:     time.Date(2018, 12, 29, 0, 0, 0, 0, time.Local),
		ReleaseDate:   time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local),
		VFXDueDate:    time.Date(2018, 9, 31, 0, 0, 0, 0, time.Local),
		OutputSize:    "1920x1080",
		LutFile:       "some/place/aces.lut",
	}
	got := ordMapFromProject(p)

	want := newOrdMap()
	want.Set("id", "TEST")
	want.Set("name", "테스트 프로젝트")
	want.Set("status", "waiting")
	want.Set("client", "레이지 픽처스")
	want.Set("director", "윤지은")
	want.Set("producer", "김한웅")
	want.Set("vfx_supervisor", "김성환")
	want.Set("vfx_manager", "조경식")
	want.Set("cg_supervisor", "김용빈")
	want.Set("crank_in", time.Date(2018, 12, 31, 7, 30, 0, 0, time.Local))
	want.Set("crank_up", time.Date(2019, 8, 31, 19, 0, 0, 0, time.Local))
	want.Set("start_date", time.Date(2018, 12, 29, 0, 0, 0, 0, time.Local))
	want.Set("release_date", time.Date(2018, 10, 1, 0, 0, 0, 0, time.Local))
	want.Set("vfx_due_date", time.Date(2018, 9, 31, 0, 0, 0, 0, time.Local))
	want.Set("output_size", "1920x1080")
	want.Set("lut_file", "some/place/aces.lut")

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}
