package dbrepo

import (
	"database/sql"
	"vigilate/internal/config"
)

//Global application configuration reference
var app *config.AppConfig

//postgresDBRepo implements the DatabaseRepo interface for PostgreSQL
type postgresDBRepo struct {
	App *config.AppConfig //Application configuration
	DB *sql.DB/ //Database connection pool
}

//NewPostgresRepo creates and returns a new PostgreSQL repository
func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	//Store app config globally for package-level access
	app  = a

	//Return repository instance
	return &postgresDBRepo{
		App: a,
		DB: conn,
	}
}