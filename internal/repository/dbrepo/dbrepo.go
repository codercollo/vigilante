package dbrepo

import (
	"database/sql"
	"vigilate/internal/config"
	"vigilate/internal/repository"
)

// Global application configuration reference
var app *config.AppConfig

// postgresDBRepo implements the DatabaseRepo interface for PostgreSQL
type postgresDBRepo struct {
	App *config.AppConfig //Application configuration
	DB  *sql.DB           //Database connection pool
}

// NewPostgresRepo creates and returns a new PostgreSQL repository
func NewPostgresRepo(Conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	app = a
	return &postgresDBRepo{
		App: a,
		DB:  Conn,
	}
}
