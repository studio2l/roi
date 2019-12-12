package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// user는 db에 저장하는 사용자 전체 정보이다.
type user struct {
	ID          string `db:"id"`
	KorName     string `db:"kor_name"`
	Name        string `db:"name"`
	Team        string `db:"team"`
	Role        string `db:"role"`
	Email       string `db:"email"`
	PhoneNumber string `db:"phone_number"`
	EntryDate   string `db:"entry_date"`

	HashedPassword string `db:"hashed_password"`

	// 설정
	CurrentShow string `db:"current_show"`
}

// User는 일반적인 사용자 정보이다.
type User struct {
	ID          string `db:"id"`
	KorName     string `db:"kor_name"`
	Name        string `db:"name"`
	Team        string `db:"team"`
	Role        string `db:"role"`
	Email       string `db:"email"`
	PhoneNumber string `db:"phone_number"`
	EntryDate   string `db:"entry_date"`
}

var CreateTableIfNotExistsUsersStmt = `CREATE TABLE IF NOT EXISTS users (
	uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	id STRING UNIQUE NOT NULL CHECK (length(id) > 0) CHECK (id NOT LIKE '% %'),
	kor_name STRING NOT NULL,
	name STRING NOT NULL,
	team STRING NOT NULL,
	role STRING NOT NULL,
	email STRING NOT NULL,
	phone_number STRING NOT NULL,
	entry_date STRING NOT NULL,
	hashed_password STRING NOT NULL,
	current_show STRING NOT NULL
)`

// AddUser는 db에 한 명의 사용자를 추가한다.
func AddUser(db *sql.DB, id, pw string) error {
	if id == "" {
		return BadRequest("need id")
	}
	// 이 이름을 가진 사용자가 이미 있는지 검사한다.
	_, err := GetUser(db, id)
	if err == nil {
		return BadRequest(fmt.Sprintf("user already exists: %s", id))
	} else if !errors.As(err, &NotFoundError{}) {
		return err
	}
	// 패스워드 해시
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashed_password := string(hashed)
	ks, is, vs, err := dbKIVs(&user{ID: id, HashedPassword: hashed_password})
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	// 사용자 생성
	stmt := fmt.Sprintf("INSERT INTO users (%s) VALUES (%s)", keys, idxs)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

func Users(db *sql.DB) ([]*User, error) {
	ks, _, _, err := dbKIVs(&User{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM users", keys)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	us := make([]*User, 0)
	for rows.Next() {
		u := &User{}
		err := scanFromRows(rows, u)
		if err != nil {
			return nil, err
		}
		us = append(us, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not scan user: %v", err)
	}
	return us, nil
}

// GetUser는 db에서 사용자를 검색한다.
// 해당 유저를 찾지 못하면 nil과 NotFound 에러를 반환한다.
func GetUser(db *sql.DB, id string) (*User, error) {
	ks, _, _, err := dbKIVs(&User{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM users WHERE id='%s'", keys, id)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, NotFound("user", id)
	}
	u := &User{}
	err = scanFromRows(rows, u)
	return u, err
}

// UserPasswordMatch는 db에 저장된 사용자의 비밀번호와 입력된 비밀번호가 같은지를 비교한다.
// 해당 사용자가 없거나, 불러오는데 에러가 나면 false와 에러를 반환한다.
func UserPasswordMatch(db *sql.DB, id, pw string) (bool, error) {
	stmt := fmt.Sprintf("SELECT hashed_password FROM users WHERE id='%s'", id)
	rows, err := db.Query(stmt)
	if err != nil {
		return false, err
	}
	ok := rows.Next()
	if !ok {
		return false, NotFound("user", id)
	}
	var hashed_password string
	rows.Scan(&hashed_password)
	err = bcrypt.CompareHashAndPassword([]byte(hashed_password), []byte(pw))
	if err != nil {
		return false, BadRequest("password not match")
	}
	return true, nil
}

// UpdateUserParam은 User에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateUser에서 사용한다.
type UpdateUserParam struct {
	KorName     string `db:"kor_name"`
	Name        string `db:"name"`
	Team        string `db:"team"`
	Role        string `db:"role"`
	Email       string `db:"email"`
	PhoneNumber string `db:"phone_number"`
	EntryDate   string `db:"entry_date"`
}

// UpdateUser는 db에 비밀번호를 제외한 사용자 필드를 업데이트 한다.
func UpdateUser(db *sql.DB, id string, u UpdateUserParam) error {
	if id == "" {
		return errors.New("empty id")
	}
	ks, is, vs, err := dbKIVs(u)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmt := fmt.Sprintf("UPDATE users SET (%s) = (%s) WHERE id='%s'", keys, idxs, id)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

type UserConfig struct {
	CurrentShow string `db:"current_show"`
}

// UpdateUserConfig는 유저의 설정 값들을 받아온다.
func GetUserConfig(db *sql.DB, id string) (*UserConfig, error) {
	if id == "" {
		return nil, errors.New("need id")
	}
	ks, _, _, err := dbKIVs(&UserConfig{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM users WHERE id='%s'", keys, id)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, NotFound("user", id)
	}
	u := &UserConfig{}
	err = scanFromRows(rows, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// UpdateUserConfig는 유저의 설정 값들을 업데이트 한다.
func UpdateUserConfig(db *sql.DB, id string, u *UserConfig) error {
	if id == "" {
		return BadRequest("need id")
	}
	if u == nil {
		return BadRequest("user config shold not nil")
	}
	ks, is, vs, err := dbKIVs(u)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmt := fmt.Sprintf("UPDATE users SET (%s) = (%s) WHERE id='%s'", keys, idxs, id)
	if _, err := db.Exec(stmt, vs...); err != nil {
		return err
	}
	return nil
}

// UpdateUserPassword는 db에 저장된 사용자 패스워드를 수정한다.
func UpdateUserPassword(db *sql.DB, id, pw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not generate hash from password: %v", err)
	}
	hashed_password := string(hashed)
	stmt := fmt.Sprintf("UPDATE users SET hashed_password=$1 WHERE id='%s'", id)
	if _, err := db.Exec(stmt, hashed_password); err != nil {
		return err
	}
	return nil
}

// DeleteUser는 해당 id의 사용자를 지운다.
// 만일 해당 아이디의 사용자가 없다면 에러를 낸다.
func DeleteUser(db *sql.DB, id string) error {
	_, err := GetUser(db, id)
	if err != nil {
		return err
	}
	stmt := fmt.Sprintf("DELETE FROM users WHERE id='%s'", id)
	if _, err := db.Exec(stmt); err != nil {
		return err
	}
	return nil
}
