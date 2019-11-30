package roi

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// User는 사용자와 관련된 정보이다.
type User struct {
	ID          string
	KorName     string
	Name        string
	Team        string
	Role        string
	Email       string
	PhoneNumber string
	EntryDate   string
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
	hashed_password STRING NOT NULL
)`

func (u *User) dbValues() []interface{} {
	if u == nil {
		u = &User{}
	}
	vals := []interface{}{
		u.ID,
		u.KorName,
		u.Name,
		u.Team,
		u.Role,
		u.Email,
		u.PhoneNumber,
		u.EntryDate,
	}
	return vals
}

func (u *User) dbValuesWithHashedPassword(hashed_password string) []interface{} {
	return append(u.dbValues(), hashed_password)
}

var UserTableKeys = []string{
	"id",
	"kor_name",
	"name",
	"team",
	"role",
	"email",
	"phone_number",
	"entry_date",
}

var UserTableKeysWithHashedPassword = append(
	UserTableKeys,
	"hashed_password",
)

var UserTableIndices = []string{
	"$1", "$2", "$3", "$4", "$5", "$6", "$7", "$8",
}

var UserTableIndicesWithHashedPassword = append(
	UserTableIndices,
	"$9",
)

// AddUser는 db에 한 명의 사용자를 추가한다.
func AddUser(db *sql.DB, id, pw string) error {
	// 이 이름을 가진 사용자가 이미 있는지 검사한다.
	exist, err := UserExist(db, id)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("user already exists: %s", id)
	}
	// 패스워드 해시
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashed_password := string(hashed)
	// 사용자 생성
	keystr := strings.Join(UserTableKeysWithHashedPassword, ", ")
	idxstr := strings.Join(UserTableIndicesWithHashedPassword, ", ")
	stmt := fmt.Sprintf("INSERT INTO users (%s) VALUES (%s)", keystr, idxstr)
	u := &User{ID: id}
	if _, err := db.Exec(stmt, u.dbValuesWithHashedPassword(hashed_password)...); err != nil {
		return err
	}
	return nil
}

// UserExist는 db에서 사용자가 존재하는지 검사한다.
func UserExist(db *sql.DB, id string) (bool, error) {
	stmt := fmt.Sprintf("SELECT id FROM users WHERE id='%s'", id)
	rows, err := db.Query(stmt)
	if err != nil {
		return false, Internal(err)
	}
	return rows.Next(), rows.Err()
}

func Users(db *sql.DB) ([]*User, error) {
	keystr := strings.Join(UserTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM users", keystr)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, Internal(err)
	}
	us := make([]*User, 0)
	for rows.Next() {
		u := &User{}
		if err := rows.Scan(&u.ID, &u.KorName, &u.Name, &u.Team, &u.Role, &u.Email, &u.PhoneNumber, &u.EntryDate); err != nil {
			return nil, Internal(err)
		}
		us = append(us, u)
	}
	if err := rows.Err(); err != nil {
		return nil, Internal(fmt.Errorf("could not scan user: %v", err))
	}
	return us, nil
}

// GetUser는 db에서 사용자를 검색한다.
// 해당 유저를 찾지 못하면 nil과 NotFound 에러를 반환한다.
func GetUser(db *sql.DB, id string) (*User, error) {
	keystr := strings.Join(UserTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM users WHERE id='%s'", keystr, id)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, Internal(err)
	}
	ok := rows.Next()
	if !ok {
		return nil, NotFound("user", id)
	}
	u := &User{}
	if err := rows.Scan(&u.ID, &u.KorName, &u.Name, &u.Team, &u.Role, &u.Email, &u.PhoneNumber, &u.EntryDate); err != nil {
		return nil, Internal(err)
	}
	return u, nil
}

// UserPasswordMatch는 db에 저장된 사용자의 비밀번호와 입력된 비밀번호가 같은지를 비교한다.
// 해당 사용자가 없거나, 불러오는데 에러가 나면 false와 에러를 반환한다.
func UserPasswordMatch(db *sql.DB, id, pw string) (bool, error) {
	stmt := fmt.Sprintf("SELECT hashed_password FROM users WHERE id='%s'", id)
	rows, err := db.Query(stmt)
	if err != nil {
		return false, Internal(err)
	}
	ok := rows.Next()
	if !ok {
		return false, NotFound("user", id)
	}
	var hashed_password string
	rows.Scan(&hashed_password)
	err = bcrypt.CompareHashAndPassword([]byte(hashed_password), []byte(pw))
	if err != nil {
		return false, Internal(err)
	}
	return true, nil
}

// UpdateUserParam은 User에서 일반적으로 업데이트 되어야 하는 멤버의 모음이다.
// UpdateUser에서 사용한다.
type UpdateUserParam struct {
	KorName     string
	Name        string
	Team        string
	Role        string
	Email       string
	PhoneNumber string
	EntryDate   string
}

func (u UpdateUserParam) keys() []string {
	return []string{
		"kor_name",
		"name",
		"team",
		"role",
		"email",
		"phone_number",
		"entry_date",
	}
}

func (u UpdateUserParam) indices() []string {
	return dbIndices(u.keys())
}

func (u UpdateUserParam) values() []interface{} {
	return []interface{}{
		u.KorName,
		u.Name,
		u.Team,
		u.Role,
		u.Email,
		u.PhoneNumber,
		u.EntryDate,
	}
}

// UpdateUser는 db에 비밀번호를 제외한 사용자 필드를 업데이트 한다.
func UpdateUser(db *sql.DB, id string, u UpdateUserParam) error {
	if id == "" {
		return errors.New("empty id")
	}
	keystr := strings.Join(u.keys(), ", ")
	idxstr := strings.Join(u.indices(), ", ")
	stmt := fmt.Sprintf("UPDATE users SET (%s) = (%s) WHERE id='%s'", keystr, idxstr, id)
	if _, err := db.Exec(stmt, u.values()...); err != nil {
		return Internal(err)
	}
	return nil
}

// UpdateUserPassword는 db에 저장된 사용자 패스워드를 수정한다.
func UpdateUserPassword(db *sql.DB, id, pw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return Internal(fmt.Errorf("could not generate hash from password: %v", err))
	}
	hashed_password := string(hashed)
	stmt := fmt.Sprintf("UPDATE users SET hashed_password=$1 WHERE id='%s'", id)
	if _, err := db.Exec(stmt, hashed_password); err != nil {
		return Internal(err)
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
		return Internal(err)
	}
	return nil
}
