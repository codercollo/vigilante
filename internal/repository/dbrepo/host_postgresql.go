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

	stmt := `
					insert into host_services (host_id, service_id, active, schedule_number, schedule_unit, 
					status, created_at, updated_at) values ($1, 1, 1, 3, 'm', 'pending', $2, $3)
	`
	_, err = m.DB.ExecContext(ctx, stmt, newID, time.Now(), time.Now())
	if err != nil {
		return newID, err
	}

	//Return the new record ID on success
	return newID, nil
}

// GetHostByID retrieves a single host record by its ID
// It also loads all associated services linked to the host
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

	//Query to retieve all services associated with the host
	query = `select
	              hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit,
	              hs.last_check, hs.status, hs.created_at, hs.updated_at,
								s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
					  from
						    host_services hs
								left join services s on (s.id = hs.service_id)
					  where
					     	host_id = $1`

	rows, err := m.DB.QueryContext(ctx, query, h.ID)
	if err != nil {
		log.Println(err)
		return h, err
	}
	defer rows.Close()

	var hostServices []models.HostService

	for rows.Next() {
		var hs models.HostService
		//Scan services and related services metadata
		err := rows.Scan(
			&hs.ID,
			&hs.HostID,
			&hs.ServiceID,
			&hs.Active,
			&hs.ScheduleNumber,
			&hs.ScheduleUnit,
			&hs.LastCheck,
			&hs.Status,
			&hs.CreatedAt,
			&hs.UpdatedAt,
			&hs.Service.ID,
			&hs.Service.ServiceName,
			&hs.Service.Active,
			&hs.Service.Icon,
			&hs.Service.CreatedAt,
			&hs.Service.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			return h, err
		}
		hostServices = append(hostServices, hs)
	}

	h.HostServices = hostServices
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

// GetHostByID retrieves all host records from the database
// For each host, it also loads and attaches all associated services
// Hosts are returned ordered alphabetically by host name
func (m *postgresDBRepo) AllHosts() ([]models.Host, error) {
	//Create context with 3-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Query to fetch all hosts ordered by host name
	query := `
	 select id, host_name, canonical_name, url, ip, ipv6, location, os,
	 active, created_at, updated_at from hosts order by host_name
	 `

	//Execute query and return a single row
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hosts []models.Host

	//Iterate through each host row
	for rows.Next() {
		var h models.Host

		//Scan database row into Host struct fields
		err = rows.Scan(
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
			log.Println(err)
			return nil, err
		}

		//Query to retrieve all services associated with the current host
		serviceQuery := `
				 select
	              hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit,
	              hs.last_check, hs.status, hs.created_at, hs.updated_at,
								s.id, s.service_name, s.active, s.icon, s.created_at, s.updated_at
					  from
						    host_services hs
								left join services s on (s.id = hs.service_id)
					  where
					     	host_id = $1`

		serviceRows, err := m.DB.QueryContext(ctx, serviceQuery, h.ID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		var hostServices []models.HostService
		//Iterate through all services linked to the current host
		for serviceRows.Next() {
			var hs models.HostService
			//Scan services record and related service metadata
			err = serviceRows.Scan(
				&hs.ID,
				&hs.HostID,
				&hs.ServiceID,
				&hs.Active,
				&hs.ScheduleNumber,
				&hs.ScheduleUnit,
				&hs.LastCheck,
				&hs.Status,
				&hs.CreatedAt,
				&hs.UpdatedAt,
				&hs.Service.ID,
				&hs.Service.ServiceName,
				&hs.Service.Active,
				&hs.Service.Icon,
				&hs.Service.CreatedAt,
				&hs.Service.UpdatedAt,
			)
			if err != nil {
				log.Println(err)
				return nil, err
			}

			hostServices = append(hostServices, hs)
		}
		//Attach services to the current host
		h.HostServices = hostServices

		//Close service rows before processing the next host
		serviceRows.Close()

		//Append fully populated host to the slice
		hosts = append(hosts, h)
	}

	//Check for iteration errors
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	//Return all hosts
	return hosts, nil
}

// UpdateHostServiceStatus updates the active status of a service for a host
func (m *postgresDBRepo) UpdateHostServiceStatus(hostID, serviceID, active int) error {
	//Create context with timeout to prevent long-running queries
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//Update active column for specific host and service
	stmt := `
	     update host_services set active = $1 where host_id = $2 and service_id = $3		
	`

	//Execute update
	_, err := m.DB.ExecContext(ctx, stmt, active, hostID, serviceID)
	if err != nil {
		return err
	}

	return nil
}

func (m *postgresDBRepo) GetAllServiceStatusCounts() (int, int, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	        select 
					   (select count(id) from host_services where active = 1 and status = 'pending') as pending,
						 (select count(id) from host_services where active = 1 and status = 'healthy') as healthy,
						 (select count(id) from host_services where active = 1 and status = 'warning') as warning,
						 (select count(id) from host_services where active = 1 and status = 'problem') as problem
	
	`

	var pending, healthy, warning, problem int

	row := m.DB.QueryRowContext(ctx, query)
	err := row.Scan(
		&pending,
		&healthy,
		&warning,
		&problem,
	)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return pending, healthy, warning, problem, nil
}

func (m *postgresDBRepo) GetServicesByStatus(status string) ([]models.HostService, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
	select 
		hs.id, hs.host_id, hs.service_id, hs.active, hs.schedule_number, hs.schedule_unit,
		hs.last_check, hs.status, hs.created_at, hs.updated_at,
		h.host_name, s.service_name
	from
		host_services hs
	left join hosts h on hs.host_id = h.id
	left join services s on hs.service_id = s.id
	where
		hs.status = $1
	and hs.active = 1
	`

	var services []models.HostService

	rows, err := m.DB.QueryContext(ctx, query, status)
	if err != nil {
		return services, err
	}
	defer rows.Close()

	for rows.Next() {
		var h models.HostService

		err := rows.Scan(
			&h.ID,
			&h.HostID,
			&h.ServiceID,
			&h.Active,
			&h.ScheduleNumber,
			&h.ScheduleUnit,
			&h.LastCheck,
			&h.Status,
			&h.CreatedAt,
			&h.UpdatedAt,
			&h.HostName,
			&h.Service.ServiceName,
		)
		if err != nil {
			return nil, err
		}

		services = append(services, h)
	}

	return services, nil
}
