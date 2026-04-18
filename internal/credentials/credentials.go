package credentials


import(
	"os"
	"errors"
	"database/sql"
	"context"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)


// Wrapper around a DB handle to allow for methods.
type CredentialDB struct {
	db    *sql.DB
}


// Stores public facing user data.
type UserData struct {
	Name    string
	Perms   uint8
	LastLog time.Time
}


// DB errors.
var ErrUserAlreadyExists = errors.New("A user with this name already exists")
var ErrUserNotExists = errors.New("No user exists with this name")
var ErrInvalidPassword = errors.New("The password is incorrect for the given user")


// StartCredentialDB starts the credential database and initializes it if necessary.
// Creates the DB if it does not exist using createStmt for initialization.
func StartCredentialDB(dbPath string, createStmt string) *CredentialDB {
	// Check if DB already exists.
	_, dbErr := os.Stat(dbPath)
	notExists := errors.Is(dbErr, os.ErrNotExist)

	// Open DB and check for errors.
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}

	// DB did not exist, initialize it.
	if notExists {
		_, err := db.Exec(createStmt)
		if err != nil {
			panic(err)
		}

		log.Info().
			Msg("Created credential DB")
	} else {
		log.Info().
			Msg("Started credential DB")
	}

	return &CredentialDB{
		db: db,
	}
}


// Close closes the credential database.
func (cdb *CredentialDB) Close() {
	log.Info().
			Msg("Closed credential DB")

	cdb.db.Close()
}


// AddUser adds the user with the given password to the database, if the user already exists an error is returned.
func (cdb *CredentialDB) AddUser(ctx context.Context, usr string, psw string, perm uint8) error {
	pswHash, err := HashPassword(psw)
	if err != nil {
		return err
	}

	query := `INSERT INTO users (name, password, permissions) VALUES (?, ?, ?)`
	// Execute query.
	if _, err = cdb.db.ExecContext(ctx, query, usr, pswHash, perm); err != nil {
		// If an error occurred check if it was a constrain violation, if so return the corresponding error.
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				return ErrUserAlreadyExists
			}
		}

		return err
	}

	return nil
}


// ValidateUser checks whether the given user exists and the given password is correct, if not returns an error.
func (cdb *CredentialDB) ValidateUser(ctx context.Context, usr string, psw string) error {
	query := `SELECT password FROM users WHERE name = ?`

	row := cdb.db.QueryRowContext(ctx, query, usr)
	var dbPsw string

	if err := row.Scan(&dbPsw); err != nil {
		// If an error occurred check if there were no returned rows, meaning there's no entry with the given name.
		if err == sql.ErrNoRows {
			return ErrUserNotExists
		}

		return err
	}

	// Check password validity.
	if !CheckPasswordHash(psw, dbPsw) {
		return ErrInvalidPassword
	}
	return nil
}


// DeleteUser deletes the given user, if it doesn't exist an error is returned.
func (cdb *CredentialDB) DeleteUser(ctx context.Context, usr string) error {
	query := `DELETE FROM users WHERE name = ?`

	res, err := cdb.db.ExecContext(ctx, query, usr)
	if err != nil {
		return err
	}
	// Check if a user was actually deleted, otherwise return an error.
	if num, _ := res.RowsAffected(); num == 0 {
		return ErrUserNotExists
	}

	return nil 
}


// UpdateUserLastLog sets the users last log time to the current time.
func (cdb *CredentialDB) UpdateUserLastLog(ctx context.Context, usr string) error {
	query := `UPDATE users SET lastLog = ? WHERE name = ?`

	res, err := cdb.db.ExecContext(ctx, query, time.Now(), usr)
	if err != nil {
		return err
	}
	// Check if a user was actually updated, otherwise return an error.
	if num, _ := res.RowsAffected(); num == 0 {
		return ErrUserNotExists
	}
	return nil 
}


// UpdateUserPassword updates the users password to the given value.
func (cdb *CredentialDB) UpdateUserPassword(ctx context.Context, usr string, newPsw string) error {
	newPswHash, err := HashPassword(newPsw)
	if err != nil {
		return err
	}

	query := `UPDATE users SET password = ? WHERE name = ?`
	res, err := cdb.db.ExecContext(ctx, query, newPswHash, usr)
	if err != nil {
		return err
	}
	// Check if a user was actually updated, otherwise return an error.
	if num, _ := res.RowsAffected(); num == 0 {
		return ErrUserNotExists
	}
	return nil 
}


// UpdateUserPermissions updates the users permissions to the given value.
func (cdb *CredentialDB) UpdateUserPermissions(ctx context.Context, usr string, perms uint8) error {
	query := `UPDATE users SET permissions = ? WHERE name = ?`
	res, err := cdb.db.ExecContext(ctx, query, perms, usr)
	if err != nil {
		return err
	}
	// Check if a user was actually updated, otherwise return an error.
	if num, _ := res.RowsAffected(); num == 0 {
		return ErrUserNotExists
	}
	return nil 
}


// GetUsers returns a list with all users public facing data.
func (cdb *CredentialDB) GetUsers(ctx context.Context) ([]UserData, error) {
	query := `SELECT name, permissions, lastLog FROM users`
	rows, err := cdb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userList []UserData
	for rows.Next() {
		var user UserData
		var nullableTime sql.NullTime

		// Get row data.
		if err := rows.Scan(&user.Name, &user.Perms, &nullableTime); err != nil {
			return nil, err
		}
		// Since time can be null that case must be handled separately.
		if nullableTime.Valid {
			user.LastLog = nullableTime.Time
		}

		userList = append(userList, user)
	}

	// Check for errors that could have occurred during iteration.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return userList, nil
}