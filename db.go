package roi

import (
	"database/sql"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/lib/pq"
)

// CreateTableIfNotExists는 db에 해당 테이블이 없을 때 추가한다.
func CreateTableIfNotExists(db *sql.DB, table string, fields []string) error {
	// id는 어느 테이블에나 꼭 들어가야 하는 항목이다.
	fields = append(
		[]string{"id UUID PRIMARY KEY DEFAULT gen_random_uuid()"},
		fields...,
	)
	field := strings.Join(fields, ", ")
	stmt := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", table, field)
	fmt.Println(stmt)
	_, err := db.Exec(stmt)
	return err
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
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	rows, err := SelectAll(db, "users", map[string]string{"userid": id})
	if err != nil {
		return err
	}
	if rows.Next() {
		return fmt.Errorf("user %s already exists", id)
	}
	var u = User{}
	u.ID = id
	u.HashedPassword = string(hashedPassword)
	m := u.toOrdMap()
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
// 반환된 User의 ID가 비어있다면 해당 유저를 찾지 못한것이다.
func GetUser(db *sql.DB, id string) (User, error) {
	stmt := fmt.Sprintf("SELECT * FROM users WHERE userid='%s'", id)
	fmt.Println(stmt)
	rows, err := db.Query(stmt)
	if err != nil {
		return User{}, err
	}
	ok := rows.Next()
	if !ok {
		return User{}, nil
	}
	var u User
	var uuid string
	if err := rows.Scan(&uuid, &u.ID, &u.HashedPassword, &u.KorName, &u.Name, &u.Team, &u.Position, &u.Email, &u.PhoneNumber, &u.EntryDate); err != nil {
		return User{}, err
	}
	return u, nil
}

// UserHasPassword는 db에 저장된 사용자의 비밀번호와 입력된 비밀번호가 같은지를 비교한다.
func UserHasPassword(db *sql.DB, id, pw string) (bool, error) {
	u, err := GetUser(db, id)
	if err != nil {
		return false, err
	}
	if u.ID == "" {
		return false, fmt.Errorf("user '%s' not exists", id)
	}
	err = bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(pw))
	if err != nil {
		return false, err
	}
	return true, nil
}

// SetUser는 db에 비밀번호를 제외한 사용자 필드를 업데이트 한다.
func SetUser(db *sql.DB, id string, u User) error {
	m := u.toOrdMap()
	// 유저의 암호는 독립된 요청에 의해서만 업데이트하기에 제외한다.
	if ok := m.Delete("hashed_password"); !ok {
		log.Fatal("user should have \"hashed_password\" key")
	}
	setstr := ""
	i := 0
	for _, k := range m.Keys() {
		if i != 0 {
			setstr += ", "
		}
		setstr += fmt.Sprintf("%s=$%d", k, i+1)
		i++
	}
	stmt := fmt.Sprintf("UPDATE users SET %s WHERE userid='%s'", setstr, id)
	fmt.Println(stmt)
	if _, err := db.Exec(stmt, m.Values()...); err != nil {
		return err
	}
	return nil
}

// SetUserPassword는 db에 저장된 사용자 패스워드를 수정한다.
func SetUserPassword(db *sql.DB, id, pw string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	stmt := "UPDATE users SET (hashed_password=$1) WHERE userid=$2"
	if _, err := db.Exec(stmt, hashedPassword, id); err != nil {
		return err
	}
	return nil
}

// SelectProject는 db에서 특정 프로젝트 정보를 부른다.
// 반환된 Project에 Code 값이 없다면 해당 프로젝트가 없다는 뜻이다.
func SelectProject(db *sql.DB, prj string) (Project, error) {
	rows, err := SelectAll(db, "projects", map[string]string{"code": prj})
	if err != nil {
		return Project{}, err
	}
	if !rows.Next() {
		return Project{}, nil
	}
	var id string
	p := Project{}
	err = rows.Scan(
		&id, &p.Code, &p.Name, &p.Status, &p.Client,
		&p.Director, &p.Producer, &p.VFXSupervisor, &p.VFXManager, &p.CrankIn,
		&p.CrankUp, &p.StartDate, &p.ReleaseDate, &p.VFXDueDate, &p.OutputSize,
		&p.LutFile,
	)
	if err != nil {
		return Project{}, err
	}
	return p, nil
}

// AddProject는 db에 프로젝트를 추가한다.
func AddProject(db *sql.DB, prj string) error {
	if _, err := db.Exec("INSERT INTO projects (code) VALUES ($1)", prj); err != nil {
		return err
	}
	// TODO: add project info, task, tracking table
	if err := CreateTableIfNotExists(db, prj+"_shots", ShotTableFields); err != nil {
		return err
	}
	return nil
}

// SelectScenes는 특정 프로젝트의 모든 씬이름을 반환한다.
func SelectScenes(db *sql.DB, prj string) ([]string, error) {
	stmt := fmt.Sprintf("SELECT DISTINCT scene FROM %s_shots", prj)
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	scenes := make([]string, 0)
	for rows.Next() {
		var sc string
		if err := rows.Scan(&sc); err != nil {
			return nil, err
		}
		scenes = append(scenes, sc)
	}
	sort.Slice(scenes, func(i int, j int) bool {
		if strings.Compare(scenes[i], scenes[j]) < 0 {
			return true
		}
		return false
	})
	return scenes, nil
}

// SearchShots는 db의 특정 프로젝트에서 검색 조건에 맞는 샷 리스트를 반환한다.
func SearchShots(db *sql.DB, prj, scene, shot, tag, status string) ([]Shot, error) {
	stmt := fmt.Sprintf("SELECT * FROM %s_shots", prj)
	i := 1
	vals := make([]interface{}, 0)
	if scene != "" {
		if i == 1 {
			stmt += " WHERE"
		} else {
			stmt += " AND"
		}
		stmt += fmt.Sprintf(" scene=$%d", i)
		vals = append(vals, scene)
		i++
	}
	if shot != "" {
		if i == 1 {
			stmt += " WHERE"
		} else {
			stmt += " AND"
		}
		stmt += fmt.Sprintf(" shot=$%d", i)
		vals = append(vals, shot)
		i++
	}
	if tag != "" {
		if i == 1 {
			stmt += " WHERE"
		} else {
			stmt += " AND"
		}
		stmt += fmt.Sprintf(" $%d::string = ANY(tags)", i)
		vals = append(vals, tag)
		i++
	}
	if status != "" {
		if i == 1 {
			stmt += " WHERE"
		} else {
			stmt += " AND"
		}
		stmt += fmt.Sprintf(" status=$%d", i)
		vals = append(vals, status)
		i++
	}
	fmt.Println(stmt)
	rows, err := db.Query(stmt, vals...)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	shots := make([]Shot, 0)
	for rows.Next() {
		var id string
		var s Shot
		if err := rows.Scan(
			&id, &s.Book, &s.Scene, &s.Name, &s.Status,
			&s.EditOrder, &s.Description, &s.CGDescription, &s.TimecodeIn, &s.TimecodeOut,
			&s.Duration, pq.Array(&s.Tags),
		); err != nil {
			return nil, fmt.Errorf("shot scan: %s", err)
		}
		shots = append(shots, s)
	}
	sort.Slice(shots, func(i int, j int) bool {
		if shots[i].Scene < shots[j].Scene {
			return true
		}
		if shots[i].Scene > shots[j].Scene {
			return false
		}
		return shots[i].Name <= shots[j].Name
	})
	return shots, nil
}

// AddShot은 db의 특정 프로젝트에 샷을 하나 추가한다.
func AddShot(db *sql.DB, prj string, s Shot) error {
	if prj == "" {
		return fmt.Errorf("project code not specified")
	}
	m := s.toOrdMap()
	keys := strings.Join(m.Keys(), ", ")
	idxs := strings.Join(pgIndices(m.Len()), ", ")
	stmt := fmt.Sprintf("INSERT INTO %s_shots (%s) VALUES (%s)", prj, keys, idxs)
	if _, err := db.Exec(stmt, m.Values()...); err != nil {
		return err
	}
	return nil
}

// FindShot은 db의 특정 프로젝트에서 샷 이름으로 해당 샷을 찾는다.
// 반환된 Shot의 Name이 비어있다면 그 이름의 샷이 없었다는 뜻이다.
// 할일 FindShot과 SelectShot은 중복의 느낌이다.
func FindShot(db *sql.DB, prj string, shot string) (Shot, error) {
	stmt := fmt.Sprintf("SELECT * FROM %s_shots WHERE shot='%s' LIMIT 1", prj, shot)
	fmt.Println(stmt)
	rows, err := db.Query(stmt)
	if err != nil {
		return Shot{}, err
	}
	ok := rows.Next()
	if !ok {
		return Shot{}, nil
	}
	var s Shot
	var id string
	if err := rows.Scan(
		&id, &s.Book, &s.Scene, &s.Name, &s.Status,
		&s.EditOrder, &s.Description, &s.CGDescription, &s.TimecodeIn, &s.TimecodeOut,
		&s.Duration, pq.Array(&s.Tags),
	); err != nil {
		return Shot{}, err
	}
	return s, nil
}

// 할일: 이 함수는 db와 상관이 없다. 파일 이름을 바꾸거나 다른 파일로 옮기자.
//
// AddThumbnail은 특정 샷의 썸네일을 등록한다.
// 썸네일은 roi안에 파일로 저장된다.
func AddThumbnail(prj, shot, thumbf string) error {
	// wrap은 AddThumbnail에서 에러가 났을 때 에러 내용에 기본적인 정보를 추가한다.
	wrap := func(err error) error {
		return fmt.Errorf("AddThumbnail: %s", err)
	}

	fi, err := os.Stat(thumbf)
	if err != nil {
		return wrap(err)
	}
	maxKB := int64(200)
	if fi.Size() > (maxKB << 10) {
		return wrap(fmt.Errorf("file size is bigger than %sKB: %s", maxKB))
	}
	from, err := os.Open(thumbf)
	if err != nil {
		return wrap(err)
	}
	defer from.Close()
	// thumbf가 지원하는 이미지 파일이 맞는지 확인한다.
	img, _, err := image.Decode(from)
	if err != nil {
		return wrap(err)
	}
	// 이미지를 png 이미지로 변경한다.
	// 파일을 부를때 일일이 파일 확장자를 검사하지 않기 위함이다.
	if err := os.MkdirAll(fmt.Sprintf("roi-userdata/thumbnail/%s", prj), 0755); err != nil {
		if !os.IsExist(err) {
			return wrap(err)
		}
	}
	to, err := os.Create(fmt.Sprintf("roi-userdata/thumbnail/%s/%s.png", prj, shot))
	if err != nil {
		return wrap(err)
	}
	defer to.Close()
	if err := png.Encode(to, img); err != nil {
		return wrap(err)
	}
	return nil
}
