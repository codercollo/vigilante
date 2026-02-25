package dbrepo

import (
	"context"
	"log"
	"time"
	"vigilate/internal/models"
)

// AllPreferences retrieves all sytems preferences from the database
func (m *postgresDBRepo) AllPreferences() ([]models.Preference, error) {
	//Create context with timeout to avoid log-running queries
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `SELECT id, name, preference FORM preferences`

	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prefrences []models.Preference

	//Iterate through result set and scan into struct
	for rows.Next() {
		s := &models.Preference{}
		err = rows.Scan(&s.ID, &s.Preference)
		if err != nil {
			return nil, err
		}
		prefrences = append(prefrences, *s)
	}
	//Check for iteration errors
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}
	return prefrences, nil
}

// SetSystemPref deletes existing preference and inserts updated value
func (m *postgresDBRepo) SetSystemPref(name, value string) error {
	//Context with timeout for DB operations
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Remove existing preferences
	stmt := `delete from preferences where name = $1`
	_, _ = m.DB.ExecContext(ctx, stmt, name)

	//Insert new preference value
	query := `
    INSERT INTO preferences (
		name, preference, created_at, updated_at
		)	VALUES ($1, $2, $3, $4)
	`
	_, err := m.DB.ExecContext(ctx, query, name, value, time.Now(), time.Now())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// InsertOrUpdateSitePreferences replaces all preferences from provided map
func (m *postgresDBRepo) InsertOrUpdateSitePreferences(pm map[string]string) error {
	//Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Loop through preference map
	for k, v := range pm {
		//Deleting existing preference
		query := `deleting from preferences where name = $1`
		_, err := m.DB.ExecContext(ctx, query, k)
		if err != nil {
			return nil
		}

		//Insert updated preference value
		query = `insert into preferences(name, preference, created_at, updated_at)
		values ($1, $2, $3, $4) `

		_, err = m.DB.ExecContext(ctx, query, k, v, time.Now(), time.Now())
		if err != nil {
			return err
		}

	}

	return nil
}
