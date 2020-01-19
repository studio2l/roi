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
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO users (%s) VALUES (%s)", keys, idxs), vs...),
	}
	return dbExec(db, stmts)
}

func Users(db *sql.DB) ([]*User, error) {
	ks, _, _, err := dbKIVs(&User{})
	if err != nil {
		return nil, err
	}
	keys := strings.Join(ks, ", ")
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM users", keys))
	us := make([]*User, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		u := &User{}
		err := scan(rows, u)
		if err != nil {
			return err
		}
		us = append(us, u)
		return nil
	})
	if err != nil {
		return nil, err
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
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM users WHERE id='%s'", keys, id))
	u := &User{}
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, u)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("user", id)
		}
		return nil, err
	}
	return u, err
}

// UserPasswordMatch는 db에 저장된 사용자의 비밀번호와 입력된 비밀번호가 같은지를 비교한다.
// 해당 사용자가 없거나, 불러오는데 에러가 나면 false와 에러를 반환한다.
func UserPasswordMatch(db *sql.DB, id, pw string) (bool, error) {
	stmt := dbStmt(fmt.Sprintf("SELECT hashed_password FROM users WHERE id='%s'", id))
	var hashed_password string
	err := dbQueryRow(db, stmt, func(row *sql.Row) error {
		return row.Scan(&hashed_password)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, NotFound("user", id)
		}
		return false, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashed_password), []byte(pw))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// UpdateUser는 db에 비밀번호를 제외한 유저 필드를 업데이트 한다.
// 이 함수를 호출하기 전 해당 유저가 존재하는지를 사용자가 검사해야한다.
func UpdateUser(db *sql.DB, id string, u *User) error {
	if u == nil {
		return fmt.Errorf("nil user")
	}
	if id == "" {
		return errors.New("empty id")
	}
	ks, is, vs, err := dbKIVs(u)
	if err != nil {
		return err
	}
	keys := strings.Join(ks, ", ")
	idxs := strings.Join(is, ", ")
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE users SET (%s) = (%s) WHERE id='%s'", keys, idxs, id), vs...),
	}
	return dbExec(db, stmts)
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
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM users WHERE id='%s'", keys, id))
	u := &UserConfig{}
	err = dbQueryRow(db, stmt, func(row *sql.Row) error {
		return scan(row, u)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NotFound("user", id)
		}
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
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE users SET (%s) = (%s) WHERE id='%s'", keys, idxs, id), vs...),
	}
	return dbExec(db, stmts)
}

// UpdateUserPassword는 db에 저장된 사용자 패스워드를 수정한다.
func UpdateUserPassword(db *sql.DB, id, pw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not generate hash from password: %v", err)
	}
	hashed_password := string(hashed)
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("UPDATE users SET hashed_password=$1 WHERE id='%s'", id), hashed_password),
	}
	return dbExec(db, stmts)
}

// DeleteUser는 해당 id의 사용자를 지운다.
// 만일 해당 아이디의 사용자가 없다면 에러를 낸다.
func DeleteUser(db *sql.DB, id string) error {
	_, err := GetUser(db, id)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("DELETE FROM users WHERE id='%s'", id)),
	}
	return dbExec(db, stmts)
}
