package connections

import (
	"database/sql"
	"fmt"
)

const (
	host    = "localhost"
	port    = 5432
	user    = "senul"
	dbname  = "iam"
	sslmode = "disable"
)

type conn struct {
	DB *sql.DB
}

func (c *conn) SetDB(db *sql.DB) {
	c.DB = db
}

var DBCon conn

func ConnectDB() (*sql.DB, error) {
	conStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s",
		host, port, user, dbname, sslmode)

	db, err := sql.Open("postgres", conStr)
	if err != nil {
		return db, err
	}

	err = db.Ping()
	return db, err
}
