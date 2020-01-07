package roi

import (
	"reflect"
	"testing"
)

func TestUser(t *testing.T) {
	u := &User{
		ID:          "kybin",
		KorName:     "김용빈",
		Name:        "kim yongbin",
		Team:        "rnd",
		Role:        "평민",
		Email:       "kybinz@gmail.com",
		PhoneNumber: "010-0000-0000",
		EntryDate:   "2018-03-02",
	}
	password := "no! this is not my password"

	db, err := testDB()
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	err = AddUser(db, u.ID, password)
	if err != nil {
		t.Fatalf("could not add user: %s", err)
	}
	err = UpdateUser(db, u.ID, u)
	if err != nil {
		t.Fatalf("could not update user: %v", err)
	}
	got, err := GetUser(db, u.ID)
	if err != nil {
		t.Fatalf("could not get user: %v", err)
	}
	if !reflect.DeepEqual(got, u) {
		t.Fatalf("user not match: got: %v, want: %v", got, u)
	}
	new_password := "this is not my password neither"
	err = UpdateUserPassword(db, u.ID, new_password)
	if err != nil {
		t.Fatalf("could not update user password: %v", err)
	}
	ok, err := UserPasswordMatch(db, u.ID, new_password)
	if err != nil {
		t.Fatalf("could not check user password match: %v", err)
	}
	if !ok {
		t.Fatalf("user password not match: %v", err)
	}
	err = DeleteUser(db, u.ID)
	if err != nil {
		t.Fatalf("could not delete user: %v", err)
	}
}
