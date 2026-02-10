package connections

import (
	"IAM-server/src/utils/env"
	"database/sql"
)

type conn struct {
	DB *sql.DB
}

func (c *conn) SetDB(db *sql.DB) {
	c.DB = db
}

var DBCon conn

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", env.DB_URI)
	if err != nil {
		return db, err
	}

	err = db.Ping()
	return db, err
}
