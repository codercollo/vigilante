package dbrepo

import (
	"context"
	"database/sql"
	"log"
	"time"
	"vigilate/internal/models"

	"golang.org/x/crypto/bcrypt"
)

//Package dbrepo implements the PostgreSQL repository for user management and authentication
//It provides functions to manage users, passwords and remember-me tokens

// AllUsers returns a slice of all active (non-deleted) users
func (m *postgresDBRepo) AllUsers() ([]*models.User, error) {
	// create a context with 3-second timeout for DB query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// SQL statement
	stmt := `SELECT id, last_name, first_name, email, user_active, created_at, updated_at FROM users
		where deleted_at is null`

	// execute query
	rows, err := m.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Ensure rows are closed after processing

	var users []*models.User

	// iterate over result set and scan each row into a User struct
	for rows.Next() {
		s := &models.User{}
		err = rows.Scan(&s.ID, &s.LastName, &s.FirstName, &s.Email, &s.UserActive, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		// Append it to the slice
		users = append(users, s)
	}

	// check for errors that occured
	if err = rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	// retrun slice of users
	return users, nil
}

// GetUserById returns a user by id
func (m *postgresDBRepo) GetUserById(id int) (models.User, error) {
	// create context with 3-second timeout for DB query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// SQL statement to select user by ID
	stmt := `SELECT id, first_name, last_name,  user_active, access_level, email, 
			created_at, updated_at
			FROM users where id = $1`
	// execute query with id parameter
	row := m.DB.QueryRowContext(ctx, stmt, id)

	var u models.User

	// Scan result into User struct
	err := row.Scan(
		&u.ID,
		&u.FirstName,
		&u.LastName,
		&u.UserActive,
		&u.AccessLevel,
		&u.Email,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		log.Println(err)
		return u, err
	}

	// return user and nil error
	return u, nil
}

// Authenticate verifies a user's email and password, returning user ID and hashed password
// if valid
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	//create context with 3-second timeout for DB query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string
	var userActive int

	// SQL  query to get user credentials and active status
	query := `
		select 
		    id, password, user_active 
		from 
			users 
		where 
			email = $1
			and deleted_at is null`

	// Execute query with email
	row := m.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&id, &hashedPassword, &userActive)
	if err == sql.ErrNoRows {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		log.Println(err)
		return 0, "", err
	}

	// compare hasheed password with provided password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", models.ErrInvalidCredentials
	} else if err != nil {
		log.Println(err)
		return 0, "", err
	}

	// check if account is active
	if userActive == 0 {
		return 0, "", models.ErrInactiveAccount
	}

	// authentication successful, return user ID and hashed password
	return id, hashedPassword, nil
}

// InsertRememberMeToken inserts a remember me token into remember_tokens for a given user ID
func (m *postgresDBRepo) InsertRememberMeToken(id int, token string) error {
	// create context with 2-second timeout for DB operation
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// SQL statement to insert user ID and token into remember_tokens table
	stmt := "insert into remember_tokens (user_id, remember_token) values ($1, $2)"
	_, err := m.DB.ExecContext(ctx, stmt, id, token)
	if err != nil {
		return err
	}

	// insert successful
	return nil
}

// DeleteToken deletes a remember me token from the database
func (m *postgresDBRepo) DeleteToken(token string) error {
	// creates context with 3-sec timeout for DB operation
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// SQL statement to delete token from remember_tokens table
	stmt := "delete from remember_tokens where remember_token = $1"
	_, err := m.DB.ExecContext(ctx, stmt, token)
	if err != nil {
		return err
	}

	// delete successful
	return nil
}

// CheckForToken verifies if a "remember me" token exists for a given user ID
func (m *postgresDBRepo) CheckForToken(id int, token string) bool {
	// create context with 3-second timeout for DB query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// SQL statement to check if token exists for the user
	stmt := "SELECT id  FROM remember_tokens where user_id = $1 and remember_token = $2"
	// execute query and scan result (id)
	row := m.DB.QueryRowContext(ctx, stmt, id, token)
	err := row.Scan(&id)

	// return true if token exists, false otherwise
	return err == nil
}

// Insert adds a new user record to the users table and returns the new ID.
func (m *postgresDBRepo) InsertUser(u models.User) (int, error) {
	// create context with 3-second timeout for DB query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword(u.Password, 12)
	if err != nil {
		return 0, err
	}

	// SQL statement to insert user and return generated ID
	stmt := `
	INSERT INTO users 
	    (
		first_name, 
		last_name, 
		email, 
		password, 
		access_level,
		user_active
		)
    VALUES($1, $2, $3, $4, $5, $6) returning id `

	var newId int

	// execute insert and scan returned ID
	err = m.DB.QueryRowContext(ctx, stmt,
		u.FirstName,
		u.LastName,
		u.Email,
		hashedPassword,
		u.AccessLevel,
		&u.UserActive).Scan(&newId)
	if err != nil {
		return 0, err
	}

	// return newly created user ID
	return newId, err
}

// UpdateUser updates an existing user record by ID
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	// create context with 3-second timeout for DB query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// SQL statement to update user fields by ID
	stmt := `
		update 
			users 
		set 
			first_name = $1, 
			last_name = $2, 
			user_active = $3, 
			email = $4, 
			access_level = $5,
			updated_at = $6
		where
			id = $7`

	// execute update with user data
	_, err := m.DB.ExecContext(ctx, stmt,
		u.FirstName,
		u.LastName,
		u.UserActive,
		u.Email,
		u.AccessLevel,
		u.UpdatedAt,
		u.ID,
	)
	if err != nil {
		log.Println(err)
		return err
	}

	// update sucessfully
	return nil
}

// DeleteUser performs a soft delete by setting deleted_at and deactivating the user
func (m *postgresDBRepo) DeleteUser(id int) error {
	// create context with 3-second timeout for DB query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// SQL statement to soft delete user
	stmt := `update users set deleted_at = $1, user_active = 0  where id = $2`

	// execute update with current timestamp and user ID
	_, err := m.DB.ExecContext(ctx, stmt, time.Now(), id)
	if err != nil {
		log.Println(err)
		return err
	}

	// delete successful
	return nil
}

// UpdatePassword resets a password
func (m *postgresDBRepo) UpdatePassword(id int, newPassword string) error {
	// create context with 3-second timeout for DB query
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Create a bcrypt hash of the plain-text password.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		log.Println(err)
		return err
	}

	// SQL statement to update user's password
	stmt := `update users set password = $1 where id = $2`
	_, err = m.DB.ExecContext(ctx, stmt, hashedPassword, id)
	if err != nil {
		log.Println(err)
		return err
	}

	// delete all remember tokens, if any
	stmt = "delete from remember_tokens where user_id = $1"
	_, err = m.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	// password update successful
	return nil
}
