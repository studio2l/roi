package roi

import (
	"database/sql"
	"log"
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
		t.Fatalf("could not initialze tables: %v", err)
	}
	err = AddUser(db, u.ID, password)
	if err != nil {
		t.Fatalf("could not add user: %s", err)
	}
	exist, err := UserExist(db, u.ID)
	if err != nil {
		t.Fatalf("could not check user exist: %v", err)
	}
	if !exist {
		t.Fatalf("add user wasn't successful")
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
	exist, err = UserExist(db, u.ID)
	if err != nil {
		t.Fatalf("could not check user exist - after delete: %v", err)
	}
	if exist {
		t.Fatalf("delete user wasn't successful")
	}
}
