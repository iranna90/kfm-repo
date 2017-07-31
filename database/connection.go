package database

import (
	_ "github.com/lib/pq"
	. "database/sql"
)

func GetDataBaseConnection() *DB {
	db, err := Open("postgres", "host=localhost port=1111 dbname=kmfdetails user=kmfadmin password=changeme001 sslmode=disable")
	if err != nil {
		panic("Error while getting the data base connection")
	}
	return db
}
