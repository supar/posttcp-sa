package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

// Create database connection
func openDB(url string) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	return
}

// Create query statment to add up spam attempts for the host or
// email address over interval.
func OpenStmt(db *sql.DB) (stmt *sql.Stmt, err error) {
	var (
		query string
	)

	query = "SELECT SUM(`spam_victims_score`) `vsum` " +
		"FROM `spammers` WHERE `client` = ? " +
		"AND `created` >= NOW() - INTERVAL ? DAY " +
		"GROUP BY `client` " +
		"HAVING (1 - POW(EXP(1), -(`vsum` / ?))) > ?"

	return db.Prepare(query)
}

// Run statment to check passed address
// Returns sum of attemps if there was found any record accoding to the statement
// otherewise 0
func isSpammer(stmt *sql.Stmt, args ...interface{}) (score int64, err error) {
	err = stmt.QueryRow(args...).Scan(&score)

	if err != nil {
		if err == sql.ErrNoRows {
			err = nil
		}
	}

	return
}
