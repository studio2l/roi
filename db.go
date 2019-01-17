package roi

import (
	"database/sql"
	"fmt"
	_ "image/jpeg"
)

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
