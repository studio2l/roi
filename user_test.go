package roi

import (
	"reflect"
	"testing"
)

func TestOrdMapFromUser(t *testing.T) {
	u := User{
		ID:          "kybin",
		KorName:     "김용빈",
		Name:        "kim yongbin",
		Team:        "rnd",
		Position:    "평민",
		Email:       "kybinz@gmail.com",
		PhoneNumber: "010-0000-0000",
		EntryDate:   "2018-03-02",
	}
	got := ordMapFromUser(u)

	want := newOrdMap()
	want.Set("userid", "kybin")
	want.Set("kor_name", "김용빈")
	want.Set("name", "kim yongbin")
	want.Set("team", "rnd")
	want.Set("position", "평민")
	want.Set("email", "kybinz@gmail.com")
	want.Set("phone_number", "010-0000-0000")
	want.Set("entry_date", "2018-03-02")

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %v, want: %v", got, want)
	}
}
