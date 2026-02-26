package dbrepo

import (
	"database/sql"
	"vigilate/internal/config"
	"vigilate/internal/repository"
)

//Package dbrepo implements the repository.DatabaseRepo interface specifically for PostgreSQL
//providing concrete database operations for users, authentication and preferences

var app *config.AppConfig

//postgresDBRepo holds PostgreSQL connection and app config
type postgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

// NewPostgresRepo initializes a new PostgreSQL repository instance
//Conn : active *sql.DB connection
//a : application config
// Returns an object implementing repository.DatabaseRepo
func NewPostgresRepo(Conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	app = a
	return &postgresDBRepo{
		App: a,
		DB:  Conn,
	}
}
