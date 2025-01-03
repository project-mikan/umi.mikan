package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // postgres driver
)

func NewDB(host string, port int, user, password, dbname string) *sql.DB {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname))
	if err != nil {
		// TODO Log出す
		log.Fatalf("failed to open db: %v", err)
	}
	return db
}
