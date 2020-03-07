package roi

import (
	"reflect"
	"testing"
	"time"
)

var testAssetA = &Asset{
	Show:          testShow.Show,
	Asset:         "woman1",
	Status:        StatusInProgress,
	Description:   "일반적인 여성",
	CGDescription: "좋아요.",
	Tags:          []string{"인간"},
	// 사이트에 이 애셋 태스크가 존재해야만 에러가 나지 않는다.
	Tasks: []string{"mod", "rig"},
}

var testAssetB = &Asset{
	Asset:         "man1",
	Show:          testShow.Show,
	Status:        StatusHold,
	Description:   "일반적인 남성",
	CGDescription: "아주 잘 했습니다.",
	Tags:          []string{"인간"},
	Tasks:         []string{"mod", "tex"},
}
var testAssetC = &Asset{
	Asset:         "char_lion_1",
	Show:          testShow.Show,
	Status:        StatusHold,
	Description:   "라이프 오브 파이급 사자",
	CGDescription: "털이 더 가늘어야 할 것 같네요.",
	Tags:          []string{"크리쳐", "사자", "털"},
	Tasks:         []string{"mod"},
}

var testAssets = []*Asset{testAssetA, testAssetB, testAssetC}

func TestAsset(t *testing.T) {
	want := testAssets

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

	for _, s := range want {
		err = AddAsset(db, s)
		if err != nil {
			t.Fatalf("could not add asset to assets table: %s", err)
		}
		got, err := GetAsset(db, s.ID())
		if err != nil {
			t.Fatalf("could not get asset from assets table: %s", err)
		}
		err = verifyAssetName(got.Asset)
		if err != nil {
			t.Fatalf("find asset with invalid id from assets table: %s", err)
		}
		if !reflect.DeepEqual(got, s) {
			t.Fatalf("got: %v, want: %v", got, s)
		}
	}

	got, err := SearchAssets(db, testShow.Show, []string{}, "", "", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search assets from assets table: %s", err)
	}
	if len(got) != 3 {
		t.Fatalf("len(got): %v, want: 3", len(got))
	}

	got, err = SearchAssets(db, testShow.Show, []string{"woman1"}, "", "", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search assets from assets table: %s", err)
	}
	if len(got) != 1 {
		t.Fatalf("len(got): %v, want: 1", len(got))
	}
	got, err = SearchAssets(db, testShow.Show, []string{}, "인간", "", "", "", "", time.Time{})
	if err != nil {
		t.Fatalf("could not search assets from assets table: %s", err)
	}
	if len(got) != 2 {
		t.Fatalf("len(got): %v, want: 2", len(got))
	}

	for _, s := range want {
		err = UpdateAsset(db, s.ID(), s)
		if err != nil {
			t.Fatalf("could not update asset: %s", err)
		}
		err = DeleteAsset(db, s.ID())
		if err != nil {
			t.Fatalf("could not delete asset from assets table: %s", err)
		}
	}
}
