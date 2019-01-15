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
	Position    string
	Email       string
	PhoneNumber string
	EntryDate   string
}

func (u *User) dbValues() []interface{} {
	if u == nil {
		u = &User{}
	}
	vals := []interface{}{
		u.ID,
		u.KorName,
		u.Name,
		u.Team,
		u.Position,
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
	"position",
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

var UserTableFields = []string{
	"uniqid UUID PRIMARY KEY DEFAULT gen_random_uuid()",
	"id STRING UNIQUE NOT NULL CHECK (length(id) > 0) CHECK (id NOT LIKE '% %')",
	"kor_name STRING NOT NULL",
	"name STRING NOT NULL",
	"team STRING NOT NULL",
	"position STRING NOT NULL",
	"email STRING NOT NULL",
	"phone_number STRING NOT NULL",
	"entry_date STRING NOT NULL",
	// hashed_password는 DB에서만 관리하고 User에는 들어가지는 않는다.
	"hashed_password STRING NOT NULL",
}

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
		return false, err
	}
	return rows.Next(), rows.Err()
}

// GetUser는 db에서 사용자를 검색한다.
// 해당 유저를 찾지 못하면 nil이 반환된다.
func GetUser(db *sql.DB, id string) (*User, error) {
	keystr := strings.Join(UserTableKeys, ", ")
	stmt := fmt.Sprintf("SELECT %s FROM users WHERE id='%s'", keystr, id)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	ok := rows.Next()
	if !ok {
		return nil, nil
	}
	u := &User{}
	if err := rows.Scan(&u.ID, &u.KorName, &u.Name, &u.Team, &u.Position, &u.Email, &u.PhoneNumber, &u.EntryDate); err != nil {
		return nil, err
	}
	return u, nil
}

// UserHasPassword는 db에 저장된 사용자의 비밀번호와 입력된 비밀번호가 같은지를 비교한다.
// 해당 사용자가 없거나, 불러오는데 에러가 나면 false와 에러를 반환한다.
func UserHasPassword(db *sql.DB, id, pw string) (bool, error) {
	stmt := fmt.Sprintf("SELECT hashed_password FROM users WHERE id='%s'", id)
	rows, err := db.Query(stmt)
	if err != nil {
		return false, err
	}
	ok := rows.Next()
	if !ok {
		return false, fmt.Errorf("user '%s' not exists", id)
	}
	var hashed_password string
	rows.Scan(&hashed_password)
	err = bcrypt.CompareHashAndPassword([]byte(hashed_password), []byte(pw))
	if err != nil {
		return false, err
	}
	return true, nil
}

// UpdateUser는 db에 비밀번호를 제외한 사용자 필드를 업데이트 한다.
func UpdateUser(db *sql.DB, id string, u *User) error {
	if id == "" {
		return errors.New("empty id")
	}
	if u == nil {
		return errors.New("nil User is invalid")
	}
	keystr := strings.Join(UserTableKeys, ", ")
	idxstr := strings.Join(UserTableIndices, ", ")
	stmt := fmt.Sprintf("UPDATE users SET (%s) = (%s) WHERE id='%s'", keystr, idxstr, id)
	if _, err := db.Exec(stmt, u.dbValues()...); err != nil {
		return err
	}
	return nil
}

// UpdateUserPassword는 db에 저장된 사용자 패스워드를 수정한다.
func UpdateUserPassword(db *sql.DB, id, pw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("could not generate hash from password: %s", err.Error())
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
	stmt := fmt.Sprintf("DELETE FROM users WHERE id='%s'", id)
	if _, err := db.Exec(stmt); err != nil {
		return err
	}
	return nil
}
