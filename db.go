package roi

import (
	"database/sql"
	"fmt"
	_ "image/jpeg"
	"reflect"
	"strconv"

	"github.com/lib/pq"
)

// InitDB는 로이 DB 및 DB유저를 생성한고 생성된 DB를 반환한다.
// 여러번 실행해도 문제되지 않는다.
// 실패하면 진행된 프로세스를 취소하고 에러를 반환한다.
func InitDB() (*sql.DB, error) {
	return initDB("postgresql://root@localhost:26257/roi?sslmode=disable")
}

func initDB(addr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", addr)
	if err != nil {
		return nil, Internal(fmt.Errorf("could not open database with root user: %w", err))
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, Internal(fmt.Errorf("could not begin a transaction: %w", err))
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨

	// 밑의 구문들은 다 여러번 실행해도 안전한 구문들이다.
	if _, err := tx.Exec("CREATE USER IF NOT EXISTS roiuser"); err != nil {
		return nil, Internal(fmt.Errorf("could not create user 'roiuser': %w", err))
	}
	if _, err := tx.Exec("CREATE DATABASE IF NOT EXISTS roi"); err != nil {
		return nil, Internal(fmt.Errorf("could not create db 'roi': %w", err))
	}
	if _, err := tx.Exec("GRANT ALL ON DATABASE roi TO roiuser"); err != nil {
		return nil, Internal(fmt.Errorf("could not grant 'roi' to 'roiuser': %w", err))
	}
	if _, err := tx.Exec(CreateTableIfNotExistsSitesStmt); err != nil {
		return nil, Internal(fmt.Errorf("could not create 'sites' table: %w", err))
	}
	if _, err := tx.Exec(CreateTableIfNotExistsShowsStmt); err != nil {
		return nil, Internal(fmt.Errorf("could not create 'projects' table: %w", err))
	}
	if _, err := tx.Exec(CreateTableIfNotExistsShotsStmt); err != nil {
		return nil, Internal(fmt.Errorf("could not create 'shots' table: %w", err))
	}
	if _, err := tx.Exec(CreateTableIfNotExistsTasksStmt); err != nil {
		return nil, Internal(fmt.Errorf("could not create 'tasks' table: %w", err))
	}
	if _, err := tx.Exec(CreateTableIfNotExistsVersionsStmt); err != nil {
		return nil, Internal(fmt.Errorf("could not create 'versions' table: %w", err))
	}
	if _, err := tx.Exec(CreateTableIfNotExistsUsersStmt); err != nil {
		return nil, Internal(fmt.Errorf("could not create 'users' table: %w", err))
	}
	err = tx.Commit()
	if err != nil {
		return nil, Internal(fmt.Errorf("could not commit transaction: %w", err))
	}
	return db, nil
}

// DB는 로이의 DB 핸들러를 반환한다. 이 함수는 이미 로이 DB와 DB 유저가 생성되어 있다고 가정한다.
func DB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgresql://roiuser@localhost:26257/roi?sslmode=disable")
	if err != nil {
		return nil, Internal(err)
	}
	return db, nil
}

// dbKVs는 임의의 타입인 v에 대해서 그 db키, 값 리스트를 반환한다.
// 참고: dbKVs는 nil 슬라이스를 빈 슬라이스로 변경한다.
func dbKVs(v interface{}) ([]string, []interface{}, error) {
	keys, err := dbKeys(v)
	if err != nil {
		return nil, nil, err
	}
	vals, err := dbValues(v)
	if err != nil {
		return nil, nil, err
	}
	return keys, vals, nil
}

// dbKeys는 임의의 타입인 v에 대해서 그 db 키 슬라이스를 반환한다.
func dbKeys(v interface{}) (keys []string, err error) {
	var typ reflect.Type
	var field reflect.StructField
	defer func() {
		if r := recover(); r != nil {
			err = Internal(fmt.Errorf("dbKeys: %v: %s.%s", r, typ.Name(), field.Name))
		}
	}()
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		// 포인터에서 스트럭트로
		rv = rv.Elem()
	}
	typ = rv.Type()
	n := typ.NumField()
	keys = make([]string, n)
	for i := 0; i < n; i++ {
		field = typ.Field(i)
		key := field.Tag.Get("db")
		if key == "" {
			panic(fmt.Errorf("no db tag value in struct"))
		}
		keys[i] = key
	}
	return keys, nil
}

// dbValues는 임의의 타입인 v에 대해서 그 값 슬라이스를 반환한다.
// 참고: dbValues는 nil 슬라이스를 빈 슬라이스로 변경한다.
func dbValues(v interface{}) (vals []interface{}, err error) {
	var typ reflect.Type
	var field reflect.Value
	defer func() {
		if r := recover(); r != nil {
			err = Internal(fmt.Errorf("dbValues: %v: %s.%s", r, typ.Name(), field.Type().Name()))
		}
	}()
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		// 포인터에서 스트럭트로
		rv = rv.Elem()
	}
	typ = rv.Type()
	n := rv.NumField()
	vals = make([]interface{}, n)
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
	return vals, nil
}

// dbIndices는 받아들인 문자열 슬라이스와 같은 길이의
// DB 인덱스 슬라이스를 생성한다. 인덱스는 1부터 시작한다.
//
// 예)
// 	dbIndices([]string{"a", "b", "c"}) => []string{"$1", "$2", "$3"}
//
func dbIndices(keys []string) []string {
	idxs := make([]string, len(keys))
	for i := range keys {
		idxs[i] = "$" + strconv.Itoa(i+1)
	}
	return idxs
}
