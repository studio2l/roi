package roi

import (
	"reflect"
	"testing"
	"time"
)

var testUnitA = &Unit{
	Show:          testShow.Show,
	Category:      "shot",
	Group:         testGroup.Group,
	Unit:          "0010",
	Status:        StatusInProgress,
	EditOrder:     10,
	Description:   "방에 우두커니 혼자 않아 있는 로이.",
	CGDescription: "조명판 들고 있는 사람이 촬영되었으니 지워주세요.",
	Tags:          []string{"로이", "리무브"},
	Assets:        []string{},
	// 사이트에 이 샷 태스크가 존재해야만 에러가 나지 않는다.
	Tasks: []string{"fx"},
	Attrs: DBStringMap{
		"timecode_in":  "00:00:00:01",
		"timecode_out": "00:00:05:12",
		"duration":     "132",
	},
}

var testUnitB = &Unit{
	Show:          testShow.Show,
	Category:      "shot",
	Group:         testGroup.Group,
	Unit:          "0020",
	Status:        StatusHold,
	EditOrder:     20,
	Description:   "고개를 돌려 창문 밖을 바라본다.",
	CGDescription: "전반적인 느낌을 어둡게 바꿔주세요.",
	Tags:          []string{"로이", "창문"},
	Assets:        []string{},
	Tasks:         []string{"lit"},
	Attrs: DBStringMap{
		"timecode_in":  "00:00:05:12",
		"timecode_out": "00:00:06:03",
		"duration":     "15",
	},
}

var testUnitC = &Unit{
	Show:          testShow.Show,
	Category:      "shot",
	Group:         testGroup.Group,
	Unit:          "0030",
	Status:        StatusHold,
	EditOrder:     30,
	Description:   "쓸쓸해 보이는 가로등",
	CGDescription: "가로등이 너무 깨끗하니 레트로 한 느낌을 살려주세요.",
	Tags:          []string{"가로등", "창문"},
	Assets:        []string{},
	Tasks:         []string{"comp"},
	Attrs: DBStringMap{
		"timecode_in":  "00:00:06:03",
		"timecode_out": "00:00:08:15",
		"duration":     "36",
	},
}

var testUnits = []*Unit{testUnitA, testUnitB, testUnitC}

func TestUnit(t *testing.T) {
	want := testUnits

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
		err = DeleteShow(db, testShow.ID())
		if err != nil {
			t.Fatalf("could not delete project: %s", err)
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
	for _, s := range want {
		err = AddUnit(db, s)
		if err != nil {
			t.Fatalf("could not add unit to units table: %s", err)
		}
		got, err := GetUnit(db, s.Show, s.Category, s.Group, s.Unit)
		if err != nil {
			t.Fatalf("could not get unit from units table: %s", err)
		}
		err = verifyUnitName(got.Unit)
		if err != nil {
			t.Fatalf("find unit with invalid id from units table: %s", err)
		}
		if !reflect.DeepEqual(got, s) {
			t.Fatalf("got: %v, want: %v", got, s)
		}
	}

	got, err := SearchUnits(db, testShow.Show, "shot", "", []string{}, "", "", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search units from units table: %s", err)
	}
	if !reflect.DeepEqual(got, want) {
		// 애셋을 테스트 중에 두 슬라이스의 순서가 다를 수 있다는 것을 발견하였다.
		// 운이 좋게 여기서 에러가 나지 않았지만 비교 방식을 바꾸어야 한다.
		// 애셋 테스트 코드처럼 갯수를 비교하거나 더 나은 비교 방식을 생각해보자.
		t.Fatalf("got: %v, want: %v", got, want)
	}

	got, err = SearchUnits(db, testShow.Show, "shot", "CG", []string{"0010"}, "", "", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search units from units table: %s", err)
	}
	want = []*Unit{testUnitA}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}
	got, err = SearchUnits(db, testShow.Show, "shot", "", []string{}, "로이", "", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search units from units table: %s", err)
	}
	want = []*Unit{testUnitA, testUnitB}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}

	for _, s := range want {
		err = UpdateUnit(db, s)
		if err != nil {
			t.Fatalf("could not update unit: %s", err)
		}
		err = DeleteUnit(db, s.Show, s.Category, s.Group, s.Unit)
		if err != nil {
			t.Fatalf("could not delete unit from units table: %s", err)
		}
	}
}
