package roi

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/lib/pq"
)

type dbItem interface {
	dbKeyValues() []KV
}

type KV struct {
	K string
	V string
}

func q(s string) string {
	s = strings.Replace(s, "'", "''", -1)
	return fmt.Sprint("'", s, "'")
}

func dbDate(t time.Time) string {
	ft := t.Format("2006-01-02")
	return "DATE " + q(ft)
}

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

func InsertInto(db *sql.DB, table string, item dbItem) error {
	keys := make([]string, 0)
	values := make([]string, 0)
	for _, kv := range item.dbKeyValues() {
		keys = append(keys, kv.K)
		values = append(values, kv.V)
	}
	keystr := strings.Join(keys, ", ")
	valuestr := strings.Join(values, ", ")
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, keystr, valuestr)
	fmt.Println(stmt)
	_, err := db.Exec(stmt)
	return err
}

func Update(db *sql.DB, table string, where string, kvs []KV) error {
	setstr := ""
	for i, kv := range kvs {
		if i != 0 {
			setstr += ", "
		}
		setstr += kv.K + "=" + kv.V
	}
	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, setstr, where)
	fmt.Println(stmt)
	_, err := db.Exec(stmt)
	return err
}

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
	if err := InsertInto(db, "users", User{ID: id, HashedPassword: string(hashedPassword)}); err != nil {
		return err
	}
	return nil
}

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

func SetUser(db *sql.DB, id string, u User) error {
	kvs := u.dbKeyValues()
	fmt.Println(kvs)
	// 유저의 암호는 독립된 요청에 의해서만 업데이트하기에 제외한다.
	find := -1
	for i, kv := range kvs {
		if kv.K == "hashed_password" {
			find = i
		}
	}
	if find == -1 {
		log.Fatal("user should have \"hashed_password\" key")
	}
	kvs = append(kvs[:find], kvs[find+1:]...)
	fmt.Println(kvs)
	if err := Update(db, "users", "userid="+q(id), kvs); err != nil {
		return err
	}
	return nil
}

func SetUserPassword(db *sql.DB, id, pw string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	kvs := []KV{{"hashed_password", q(string(hashedPassword))}}
	if err := Update(db, "users", "userid="+q(id), kvs); err != nil {
		return err
	}
	return nil
}

func AddProject(db *sql.DB, prj string) error {
	if err := InsertInto(db, "projects", Project{Code: prj}); err != nil {
		return err
	}
	// TODO: add project info, task, tracking table
	if err := CreateTableIfNotExists(db, prj+"_shots", ShotTableFields); err != nil {
		return err
	}
	return nil
}

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

func SelectShots(db *sql.DB, prj string, where map[string]string) ([]Shot, error) {
	rows, err := SelectAll(db, prj+"_shots", where)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	shots := make([]Shot, 0)
	for rows.Next() {
		var id string
		var s Shot
		if err := rows.Scan(&id, &s.Book, &s.Scene, &s.Name, &s.Status, &s.Description, &s.CGDescription, &s.TimecodeIn, &s.TimecodeOut); err != nil {
			return nil, err
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

func AddShot(db *sql.DB, prj string, s Shot) error {
	if prj == "" {
		return fmt.Errorf("project code not specified")
	}
	if err := InsertInto(db, prj+"_shots", s); err != nil {
		return err
	}
	return nil
}

func FindShot(db *sql.DB, prj string, s string) (Shot, error) {
	stmt := fmt.Sprintf("SELECT * FROM %s_shots WHERE shot='%s' LIMIT 1", prj, s)
	fmt.Println(stmt)
	rows, err := db.Query(stmt)
	if err != nil {
		return Shot{}, err
	}
	ok := rows.Next()
	if !ok {
		return Shot{}, nil
	}
	var shot Shot
	var id string
	if err := rows.Scan(&id, &shot.Book, &shot.Scene, &shot.Name, &shot.Status, &shot.Description, &shot.CGDescription, &shot.TimecodeIn, &shot.TimecodeOut); err != nil {
		return Shot{}, err
	}
	return shot, nil
}
