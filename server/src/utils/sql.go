package utils

import "database/sql"

func HadleNullSqlString(value *sql.NullString) string {
	if value.Valid {
		return value.String
	}
	return ""
}
