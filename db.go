package roi

import (
	"database/sql"
	"fmt"
	_ "image/jpeg"
	"log"
	"strconv"
)

// InitDB는 로이 DB 및 DB유저를 생성한다.
// 여러번 실행해도 문제되지 않는다.
// 실패하면 진행된 프로세스를 취소하고 에러를 반환한다.
func InitDB() error {
	db, err := sql.Open("postgres", "postgresql://root@localhost:26257/roi?sslmode=disable")
	if err != nil {
		return fmt.Errorf("could not the database with root user: %v", err)
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("could not begin a transaction: %v", err)
	}
	defer tx.Rollback() // 트랜잭션이 완료되지 않았을 때만 실행됨

	// 밑의 구문들은 다 여러번 실행해도 안전한 구문들이다.
	if _, err := tx.Exec("CREATE USER IF NOT EXISTS roiuser"); err != nil {
		log.Fatal("could not create user 'roiuser': ", err)
	}
	if _, err := tx.Exec("CREATE DATABASE IF NOT EXISTS roi"); err != nil {
		log.Fatal("could not create db 'roi': ", err)
	}
	if _, err := tx.Exec("GRANT ALL ON DATABASE roi TO roiuser"); err != nil {
		log.Fatal("could not grant 'roi' to 'roiuser': ", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit the transaction: %v", err)
	}
	return nil
}

// DB는 로이의 DB 핸들러를 반환한다. 이 함수는 이미 로이 DB가 생성되어 있다고 가정한다.
func DB() (*sql.DB, error) {
	return sql.Open("postgres", "postgresql://roiuser@localhost:26257/roi?sslmode=disable")
}

// InitTables는 DB에 로이와 관련된 테이블을 생성한다.
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
	if _, err := tx.Exec(CreateTableIfNotExistsTasksStmt); err != nil {
		return fmt.Errorf("could not create 'tasks' table: %v", err)
	}
	if _, err := tx.Exec(CreateTableIfNotExistsVersionsStmt); err != nil {
		return fmt.Errorf("could not create 'versions' table: %v", err)
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
