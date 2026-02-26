package dbrepo

import (
	"context"
	"log"
	"time"
	"vigilate/internal/models"
)

//Package dbrepo provides the PostgreSQL implementation of the DatabaseRepo interface
//handling system preferences and site configuration storage and retrievaal

// AllPreferences retrieves all preferences from the database
func (m *postgresDBRepo) AllPreferences() ([]models.Preference, error) {
	// set timeout context for query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := "SELECT id, name, preference FROM preferences"

	// execute query
	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var preferences []models.Preference

	// iterate over rows and scan into Preference struct
	for rows.Next() {
		s := &models.Preference{}
		err = rows.Scan(&s.ID, &s.Name, &s.Preference)
		if err != nil {
			return nil, err
		}
		preferences = append(preferences, *s)
	}

	// check for errors during iteration
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return preferences, nil
}

// SetSystemPref deletes and sets a single system preference
func (m *postgresDBRepo) SetSystemPref(name, value string) error {
	// set timeout context for query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//remove existing existing preference
	stmt := `delete from preferences where name = $1`
	_, _ = m.DB.ExecContext(ctx, stmt, name)

	//insert new preference
	query := `
		INSERT INTO preferences (
			  	name, preference, created_at, updated_at
			  ) VALUES ($1, $2, $3, $4)`

	_, err := m.DB.ExecContext(ctx, query, name, value, time.Now(), time.Now())
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// InsertOrUpdateSitePreferences inserts or updates all site prefs from map
func (m *postgresDBRepo) InsertOrUpdateSitePreferences(pm map[string]string) error {
	// set timeout context for queries
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	for k, v := range pm {
		// delete existing preferance
		query := `delete from preferences where name = $1`

		_, err := m.DB.ExecContext(ctx, query, k)
		if err != nil {
			return err
		}

		// insert new preference
		query = `insert into preferences (name, preference, created_at, updated_at)
			values ($1, $2, $3, $4)`

		_, err = m.DB.ExecContext(ctx, query, k, v, time.Now(), time.Now())
		if err != nil {
			return err
		}
	}

	return nil
}
