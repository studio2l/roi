package roi

import (
	"reflect"
	"testing"
)

var testGroup = &Group{
	Show:     testShow.Show,
	Category: "shot",
	Group:    "CG",
	Notes:    "hi!",
	Attrs: DBStringMap{
		"lut": "some/other/lut.cube",
	},
}

func TestGroup(t *testing.T) {
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

	s := testGroup
	err = AddGroup(db, s)
	if err != nil {
		t.Fatalf("could not add group to groups table: %s", err)
	}
	got, err := GetGroup(db, s.Show, s.Category, s.Group)
	if err != nil {
		t.Fatalf("could not get group from groups table: %s", err)
	}
	if !reflect.DeepEqual(got, s) {
		t.Fatalf("got: %v, want: %v", got, s)
	}
	err = UpdateGroup(db, s)
	if err != nil {
		t.Fatalf("could not update group: %s", err)
	}
	err = DeleteGroup(db, s.Show, s.Category, s.Group)
	if err != nil {
		t.Fatalf("could not delete group from groups table: %s", err)
	}
}
