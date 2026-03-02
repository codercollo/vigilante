package dbrepo

import (
	"context"
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
