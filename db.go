package roi

import (
	"database/sql"
	"errors"
	"fmt"
	_ "image/jpeg"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func InitTables(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨
	if _, err := tx.Exec(CreateTableIfNotExistsProjectsStmt); err != nil {
		return fmt.Errorf("could not create 'projects' table: %v", err)
	}
	if _, err := tx.Exec(CreateTableIfNotExistsShotsStmt); err != nil {
		return fmt.Errorf("could not create 'shots' table: %v", err)
	}
	if _, err := tx.Exec(CreateTableIfNotExistsUsersStmt); err != nil {
		return fmt.Errorf("could not create 'users' table: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit the transaction: %v", err)
	}
	return nil
}

// SelectAll은 특정 db 테이블의 모든 열을 검색하여 *sql.Rows 형태로 반환한다.
func SelectAll(db *sql.DB, table string, where map[string]string) (*sql.Rows, error) {
	stmt := fmt.Sprintf("SELECT * FROM %s", table)
	if len(where) != 0 {
		wheres := ""
		for k, v := range where {
			if wheres != "" {
				wheres += " AND "
			}
			wheres += fmt.Sprintf("(%s = '%s')", k, v)
		}
		stmt += " WHERE " + wheres
	}
	fmt.Println(stmt)
	return db.Query(stmt)
}

// pgIndices는 "$1" 부터 "$n"까지의 문자열 슬라이스를 반환한다.
// 이는 postgres에 대한 db.Exec나 db.Query를 위한 질의문을 만들때 유용하게 쓰인다.
func pgIndices(n int) []string {
	if n <= 0 {
		return []string{}
	}
	idxs := make([]string, n)
	for i := 0; i < n; i++ {
		idxs[i] = fmt.Sprintf("$%d", i+1)
	}
	return idxs
}

// AddUser는 db에 한 명의 사용자를 추가한다.
func AddUser(db *sql.DB, id, pw string) error {
	// 이 이름을 가진 사용자가 이미 있는지 검사한다.
	rows, err := SelectAll(db, "users", map[string]string{"id": id})
	if err != nil {
		return err
	}
	if rows.Next() {
		return fmt.Errorf("user %s already exists", id)
	}
	// 패스워드 해시
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashed_password := string(hashed)
	// 사용자 생성
	m := NewUserMap(id, hashed_password)
	keystr := strings.Join(m.Keys(), ", ")
	idxstr := strings.Join(pgIndices(m.Len()), ", ")
	stmt := fmt.Sprintf("INSERT INTO users (%s) VALUES (%s)", keystr, idxstr)
	fmt.Println(stmt)
	if _, err := db.Exec(stmt, m.Values()...); err != nil {
		return err
	}
	return nil
}

// GetUser는 db에서 사용자를 검색한다.
// 해당 유저를 찾지 못하면 nil이 반환된다.
func GetUser(db *sql.DB, id string) (*User, error) {
	stmt := "SELECT id, kor_name, name, team, position, email, phone_number, entry_date FROM users WHERE id=$1"
	fmt.Println(stmt)
	rows, err := db.Query(stmt, id)
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
	stmt := "SELECT hashed_password FROM users WHERE id=$1"
	fmt.Println(stmt)
	rows, err := db.Query(stmt, id)
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
	if u == nil {
		return errors.New("nil User is invalid")
	}
	m := ordMapFromUser(u)
	setstr := ""
	i := 0
	for _, k := range m.Keys() {
		if i != 0 {
			setstr += ", "
		}
		setstr += fmt.Sprintf("%s=$%d", k, i+1)
		i++
	}
	stmt := fmt.Sprintf("UPDATE users SET %s WHERE id='%s'", setstr, id)
	fmt.Println(stmt)
	if _, err := db.Exec(stmt, m.Values()...); err != nil {
		return err
	}
	return nil
}

// UpdateUserPassword는 db에 저장된 사용자 패스워드를 수정한다.
func UpdateUserPassword(db *sql.DB, id, pw string) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	hashed_password := string(hashed)
	stmt := "UPDATE users SET hashed_password=$1 WHERE id=$2"
	fmt.Println(stmt)
	if _, err := db.Exec(stmt, hashed_password, id); err != nil {
		return err
	}
	return nil
}
