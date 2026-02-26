package driver

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//Package driver handles database connection pooling for the application

// DB wraps the sql.DB connection pool for PostgreSQL
type DB struct {
	SQL *sql.DB
}

// dbConn is a global instance of DB used throughout the application
var dbConn = &DB{}

// Database connection pool settings
const maxOpenDbConn = 25
const maxIdleDbConn = 25
const maxDbLifetime = 5 * time.Minute

// ConnectPostgres initializes and configures the Postgres connection pool
// sets up connection pooling and verifies connectivity
func ConnectPostgres(dsn string) (*DB, error) {
	//Open connection using pgx driver
	d, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	//Configure pool limits
	d.SetMaxOpenConns(maxIdleDbConn)
	d.SetMaxIdleConns(maxIdleDbConn)
	d.SetConnMaxIdleTime(maxDbLifetime)

	dbConn.SQL = d

	//Verify database connection
	err = testDB(err, d)
	return dbConn, err
}

// testDB checks if database is reachable
func testDB(err error, d *sql.DB) error {
	//Ping database
	err = d.Ping()
	if err != nil {
		fmt.Println("Error!", err)
	} else {
		log.Println("*** Pinged database successfully! ***")
	}
	return err
}
