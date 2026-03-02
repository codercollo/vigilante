package dbrepo

import (
	"context"
	"errors"
	"log"
	"time"
	"vigilate/internal/models"
)

// InsertHost inserts a new host record into the database and returns the newly created host ID
func (m *postgresDBRepo) InsertHost(h models.Host) (int, error) {
	//Create a context with 3-sec timeout prevents long running queries
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//SQL insert query
	query := `insert into hosts(
    host_name,
    canonical_name,
    url,
    ip,
    ipv6,
    location,
    os,
    active,
    created_at,
    updated_at
) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
returning id`

	var newID int

	//Execute query and scan returned id into newID
	err := m.DB.QueryRowContext(ctx, query,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		time.Now(),
	).Scan(&newID)

	//Log and return error if insert fails
	if err != nil {
		log.Println(err)
		return newID, err
	}

	//Return the new record ID on success
	return newID, nil
}

// GetHostByID retrieves a single host record by its ID
func (m *postgresDBRepo) GetHostByID(id int) (models.Host, error) {
	//Create a context with 3-sec timeout prevents long running queries
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//SQL query to select host by ID
	query := `
	      select id,
				host_name,
				canonical_name,
				url,
				ip,
				ipv6,
				location,
				os,
				active,
				created_at,
				updated_at
				from hosts where id = $1
	`
	//Execute query with provided ID
	row := m.DB.QueryRowContext(ctx, query, id)

	var h models.Host

	// Scan returned row into Host struct fields
	err := row.Scan(
		&h.ID,
		&h.HostName,
		&h.CanonicalName,
		&h.URL,
		&h.IP,
		&h.IPV6,
		&h.Location,
		&h.OS,
		&h.Active,
		&h.CreatedAt,
		&h.UpdatedAt,
	)
	if err != nil {
		return h, err
	}

	//Return populated Host struct
	return h, nil
}

func (m *postgresDBRepo) UpdateHost(h models.Host) error {
	//Create a context with 3-sec timeout prevents long running queries
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//SQL update statement
	stmt := `
		update hosts set
			host_name = $1,
			canonical_name = $2,
			url = $3,
			ip = $4,
			ipv6 = $5,
			location = $6,
			os = $7,
			active = $8,
			updated_at = $9
		where id = $10
	`

	//Execute update with struct values
	result, err := m.DB.ExecContext(ctx, stmt,
		h.HostName,
		h.CanonicalName,
		h.URL,
		h.IP,
		h.IPV6,
		h.Location,
		h.OS,
		h.Active,
		time.Now(),
		h.ID,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	//Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	//If no rows were updated, the ID likely does not exist
	if rowsAffected == 0 {
		return errors.New("no host found to update")
	}

	return nil
}
