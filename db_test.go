package main

import (
	"database/sql"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	"testing"
)

type Tmock struct {
	addr     string
	interval int
	coef     float64
	result   interface{}
	expect   int64
}

var t_mock = []Tmock{
	Tmock{
		addr:     "any.dot.com",
		interval: 60,
		coef:     0.22,
		result:   1,
		expect:   1,
	},
	Tmock{
		addr:     "1.1.1.1",
		interval: 60,
		coef:     0.22,
		result:   nil,
		expect:   0,
	},
}

func InitDBMock(t *testing.T) (db *sql.DB, mock sqlmock.Sqlmock) {
	var (
		err error
	)

	// open database stub
	db, mock, err = sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	return
}

func Test_isSpammer(t *testing.T) {
	var (
		score int64
		err   error
		stmt  *sql.Stmt
	)

	db, mock := InitDBMock(t)
	defer db.Close()

	q := mock.ExpectPrepare(
		"^SELECT SUM(.+) `vsum` FROM (.+) WHERE (.+) " +
			"GROUP BY (.+) HAVING " +
			`\(1 - POW\(EXP\(1\), -\(` + "`vsum` /" + ` \?\)\)\) > \?`)

	if stmt, err = OpenStmt(db); err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	// Fill expectations
	for _, m := range t_mock {
		qr := q.ExpectQuery().
			WithArgs(m.addr, m.interval, m.interval, m.coef)

		if m.result != nil {
			qr.WillReturnRows(sqlmock.NewRows([]string{"vsum"}).AddRow(m.result))
		} else {
			qr.WillReturnRows(sqlmock.NewRows([]string{"vsum"}))
		}
	}

	// Run check
	for _, m := range t_mock {
		if score, err = isSpammer(stmt, m.addr, m.interval, m.interval, m.coef); err != nil {
			t.Error(err)
		} else {
			if score != m.expect {
				t.Errorf("Expected %v, but got %v (%v)", m.expect, score, m)
			}
		}
	}

	// we make sure that all expectations were met
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("There were unfulfilled expections: %s", err.Error())
	}
}
