package roi

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	_ "image/jpeg"
	"reflect"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

// InitDB는 로이 DB 및 DB유저를 생성한고 생성된 DB를 반환한다.
// 여러번 실행해도 문제되지 않는다.
// 실패하면 진행된 프로세스를 취소하고 에러를 반환한다.
func InitDB(addr, ca, cert, key string) (*sql.DB, error) {
	return initDB(fmt.Sprintf("postgresql://root@%s/roi?sslrootcert=%s&sslcert=%s&sslkey=%s&sslmode=verify-full", addr, ca, cert, key))
}

func initDB(url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("could not open database with root user: %w", err)
	}
	// 아래 구문들은 다 여러번 실행해도 안전한 구문들이다.
	stmts := []dbStatement{
		dbStmt("CREATE USER IF NOT EXISTS roiuser"),
		dbStmt("CREATE DATABASE IF NOT EXISTS roi"),
		dbStmt("GRANT ALL ON DATABASE roi TO roiuser"),
		dbStmt(CreateTableIfNotExistsSitesStmt),
		dbStmt(CreateTableIfNotExistsShowsStmt),
		dbStmt(CreateTableIfNotExistsGroupsStmt),
		dbStmt(CreateTableIfNotExistsUnitsStmt),
		dbStmt(CreateTableIfNotExistsTasksStmt),
		dbStmt(CreateTableIfNotExistsVersionsStmt),
		dbStmt(CreateTableIfNotExistsReviewsStmt),
		dbStmt(CreateTableIfNotExistsUsersStmt),
	}
	err = dbExec(db, stmts)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// DB는 로이의 DB 핸들러를 반환한다. 이 함수는 이미 로이 DB와 DB 유저가 생성되어 있다고 가정한다.
func DB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgresql://roiuser@localhost:26257/roi?sslmode=disable")
	if err != nil {
		return nil, err
	}
	return db, nil
}

// dbStatement는 db 실행문과 그 안의 $n 인덱스를 대체할 값들이다.
type dbStatement struct {
	s  string
	vs []interface{}
}

// dbStmt는 dbStatement를 생성한다.
func dbStmt(s string, vs ...interface{}) dbStatement {
	return dbStatement{
		s:  s,
		vs: vs,
	}
}

// dbQuery는 db에서 원하는 열을 검색한 후 각 열에 대해서 scanFn을 실행한다.
func dbQuery(db *sql.DB, stmt dbStatement, scanFn func(*sql.Rows) error) error {
	rows, err := db.Query(stmt.s, stmt.vs...)
	if err != nil {
		return fmt.Errorf("dbQuery: %q %v: %w", stmt.s, stmt.vs, err)
	}
	defer rows.Close()
	for rows.Next() {
		err := scanFn(rows)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return fmt.Errorf("dbQuery: %q %v: %w", stmt.s, stmt.vs, err)
	}
	return nil
}

// dbQuery는 db에 원하는 하나의 열을 검색하고 해당 열에 대해 scanFn을 실행한다.
func dbQueryRow(db *sql.DB, stmt dbStatement, scanFn func(*sql.Row) error) error {
	row := db.QueryRow(stmt.s, stmt.vs...)
	err := scanFn(row)
	if err != nil {
		return fmt.Errorf("dbQueryRow: %q %v: %w", stmt.s, stmt.vs, err)
	}
	return nil
}

// dbExec는 여러 dbStatement를 한번의 트랜잭션으로 처리한다.
// 모든 명령이 다 성공적으로 실행되었을때만 db에 그 결과가 저장된다.
func dbExec(db *sql.DB, stmts []dbStatement) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, stmt := range stmts {
		_, err := tx.Exec(stmt.s, stmt.vs...)
		if err != nil {
			return fmt.Errorf("dbExec: %q %v: %w", stmt.s, stmt.vs, err)
		}
	}
	return tx.Commit()
}

// dbKeys는 임의의 타입인 v에 대해서 그 db 키 슬라이스를 반환한다.
func dbKeys(v interface{}) []string {
	var typ reflect.Type
	var field reflect.StructField
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		// 포인터에서 스트럭트로
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic("only accept struct")
	}
	typ = rv.Type()
	n := typ.NumField()
	keys := make([]string, n)
	for i := 0; i < n; i++ {
		field = typ.Field(i)
		key := field.Tag.Get("db")
		if key == "" {
			panic("no db tag value in struct")
		}
		keys[i] = key
	}
	return keys
}

// dbVals는 임의의 타입인 v에 대해서 그 값 슬라이스를 반환한다.
// 참고: dbValues는 nil 슬라이스를 빈 슬라이스로 변경한다.
func dbVals(v interface{}) []interface{} {
	var field reflect.Value
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		// 포인터에서 스트럭트로
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic("only accept struct")
	}
	n := rv.NumField()
	vals := make([]interface{}, n)
	for i := 0; i < n; i++ {
		field = rv.Field(i)
		fv := field.Interface()
		if field.Kind() == reflect.Slice {
			// 현재로써는 임의의 슬라이스를 DB에 넣을때 pq.Array의 힘을 빌린다.
			if field.IsNil() {
				// roi는 DB에 nil을 사용하지 않는다.
				fv = pq.Array(reflect.MakeSlice(field.Type(), 0, 0).Interface())
			} else {
				fv = pq.Array(field.Interface())
			}
		}
		vals[i] = fv
	}
	return vals
}

// dbAddrs는 임의의 타입인 v에 대해서 그 멤버들의 포인터 슬라이스를 반환한다.
func dbAddrs(v interface{}) []interface{} {
	var field reflect.Value
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		// 포인터에서 스트럭트로
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		panic("only accept struct")
	}
	n := rv.NumField()
	addrs := make([]interface{}, n)
	for i := 0; i < n; i++ {
		field = rv.Field(i)
		fv := field.Addr().Interface()
		if field.Kind() == reflect.Slice {
			fv = pq.Array(fv)
		}
		addrs[i] = fv
	}
	return addrs
}

// dbIdxs는 임의의 타입인 v에 대해서 그 인덱스 슬라이스를 반환한다.
// 참고: 이 함수가 효율적이진 않지만 프로그램 시작시에 항목당 한번만
// 실행되기 때문에 문제가 되진 않는다.
func dbIdxs(v interface{}) []string {
	keys := dbKeys(v)
	idxs := make([]string, len(keys))
	for i := range keys {
		idxs[i] = "$" + strconv.Itoa(i+1)
	}
	return idxs
}

// scanner는 sql.Row 또는 sql.Rows이다.
type scanner interface {
	Scan(dest ...interface{}) error
}

// scan은 scanner에서 다음 열을 검색해 그 정보를 스트럭트인 v의 각 필드에 넣어준다.
// 만일 v가 스트럭트가 아니거나 스캔중 문제가 생겼다면 에러를 반환한다.
func scan(s scanner, v interface{}) error {
	addrs := dbAddrs(v)
	return s.Scan(addrs...)
}

// DBStringMap은 db에 저장될 때 문자열로 변경되는 맵이다. 키와 값이 모두 문자열이다.
//
// Value와 Scan 메소드의 리시버가 하나는 값, 하나는 포인터로 다른 것이 이상해 보이겠지만
// 이렇게 해야지만 정상 작동하였다.
type DBStringMap map[string]string

// Value는 db에 저장될 값이다.
func (m DBStringMap) Value() (driver.Value, error) {
	val := ""
	i := 0
	for k, v := range m {
		if i != 0 {
			val += "\n"
		}
		val += k + ": " + v
		i++
	}
	return val, nil
}

// Scan은 db의 문자열 값을 맵으로 가져온다.
func (m *DBStringMap) Scan(src interface{}) error {
	val, ok := src.(string)
	if !ok {
		return fmt.Errorf("DBStringMap: src should be string")
	}
	*m = make(DBStringMap, 0)
	for _, l := range strings.Split(val, "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		kv := strings.SplitN(l, ": ", 2)
		if len(kv) != 2 {
			return fmt.Errorf("DBStringMap: line need a colon to split key and value: got %s", l)
		}
		k := kv[0]
		v := kv[1]
		(*m)[k] = v
	}
	return nil
}
