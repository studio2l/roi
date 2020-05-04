package roi

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// CreateTableIfNotExistReviewsStmt는 DB에 reviews 테이블을 생성하는 sql 구문이다.
// 테이블은 타입보다 많은 정보를 담고 있을수도 있다.
var CreateTableIfNotExistsReviewsStmt = `CREATE TABLE IF NOT EXISTS reviews (
	show STRING NOT NULL CHECK (length(show) > 0) CHECK (show NOT LIKE '% %'),
	category STRING NOT NULL CHECK (length(category) > 0) CHECK (category NOT LIKE '% %'),
	unit STRING NOT NULL CHECK (length(unit) > 0) CHECK (unit NOT LIKE '% %'),
	task STRING NOT NULL CHECK (length(task) > 0) CHECK (task NOT LIKE '% %'),
	version STRING NOT NULL CHECK (length(version) > 0),
	created TIMESTAMPTZ NOT NULL,
	messenger STRING NOT NULL CHECK (length(messenger) > 0),
	reviewer STRING NOT NULL CHECK (length(reviewer) > 0),
	msg STRING NOT NULL,
	status STRING NOT NULL,
	CONSTRAINT reviews_pk PRIMARY KEY (show, category, unit, task, version, created)
)`

// Review는 버전의 상태를 변경하거나 수정을 위한 메시지를 남긴다.
type Review struct {
	Show     string    `db:"show"`
	Category string    `db:"category"`
	Unit     string    `db:"unit"`
	Task     string    `db:"task"`
	Version  string    `db:"version"`
	Created  time.Time `db:"created"` // 작성 시간; 항목 생성시 자동으로 입력된다.

	Reviewer  string `db:"reviewer"`  // 리뷰 한 사람의 아이디
	Messenger string `db:"messenger"` // 메시지를 작성한 사람의 아이디
	Msg       string `db:"msg"`       // 리뷰 내용
	Status    Status `db:"status"`    // 이 상태로 변경되었음
}

var reviewDBKey string = strings.Join(dbKeys(&Review{}), ", ")
var reviewDBIdx string = strings.Join(dbIdxs(&Review{}), ", ")
var _ []interface{} = dbVals(&Review{})

// verifyReview는 받아들인 리뷰가 유효하지 않다면 에러를 반환한다.
// 필요하다면 db의 정보와 비교하거나 유효성 확보를 위해 정보를 수정한다.
func verifyReview(db *sql.DB, r *Review) error {
	if r == nil {
		return fmt.Errorf("nil review")
	}
	err := verifyShowName(r.Show)
	if err != nil {
		return err
	}
	err = verifyCategoryName(r.Category)
	if err != nil {
		return err
	}
	err = verifyUnitName(r.Unit)
	if err != nil {
		return err
	}
	err = verifyTaskName(r.Task)
	if err != nil {
		return err
	}
	err = verifyVersionName(r.Version)
	if err != nil {
		return err
	}
	if r.Messenger == "" {
		return BadRequest(fmt.Sprintf("messenger should specified"))
	}
	// 할일: 리뷰 전달자가 수퍼바이저 또는 PM인지 확인한다.

	// 할일: Reviewer와 관련한 처리

	// 리뷰가 태스크의 상태를 변경한다면 그 상태에 대한 문자열,
	// 변경하지 않는다면 빈 문자열이다.
	if r.Status != "" {
		err = verifyReviewStatus(r.Status)
		if err != nil {
			return err
		}
	}
	r.Created = time.Now()
	return nil
}

// AddReview는 db의 특정 버전에 리뷰를 하나 추가한다.
func AddReview(db *sql.DB, r *Review) error {
	err := verifyReview(db, r)
	if err != nil {
		return err
	}
	// 부모가 있는지 검사
	_, err = GetVersion(db, r.Show+"/"+r.Category+"/"+r.Unit+"/"+r.Task+"/"+r.Version)
	if err != nil {
		return err
	}
	stmts := []dbStatement{
		dbStmt(fmt.Sprintf("INSERT INTO reviews (%s) VALUES (%s)", reviewDBKey, reviewDBIdx), dbVals(r)...),
	}
	return dbExec(db, stmts)
}

// VersionReviews는 해당 버전의 리뷰들을 반환한다.
func VersionReviews(db *sql.DB, id string) ([]*Review, error) {
	show, ctg, unit, task, ver, err := SplitVersionID(id)
	if err != nil {
		return nil, err
	}
	_, err = GetVersion(db, id)
	if err != nil {
		return nil, err
	}
	stmt := dbStmt(fmt.Sprintf("SELECT %s FROM reviews WHERE show=$1 AND category=$2 AND unit=$3 AND task=$4 AND version=$5", reviewDBKey), show, ctg, unit, task, ver)
	reviews := make([]*Review, 0)
	err = dbQuery(db, stmt, func(rows *sql.Rows) error {
		r := &Review{}
		err := scan(rows, r)
		if err != nil {
			return err
		}
		reviews = append(reviews, r)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return reviews, nil
}
