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
	if _, err := tx.Exec(CreateTableIfNotExistsTasksStmt); err != nil {
		return fmt.Errorf("could not create 'tasks' table: %v", err)
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
