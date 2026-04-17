package credentials


import(
	"os"
	"errors"
	"database/sql"
	"context"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)


type CredentialDB struct {
	db    *sql.DB
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
			Msg("Started credential DB")
	} else {
		log.Info().
			Msg("Created credential DB")
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
	_, err = cdb.db.ExecContext(ctx, query, usr, pswHash, perm)
	if err != nil {
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


// GetUser checks whether the given user exists and the given password is correct, if so it returns that users permissions,
// otherwise it returns an error.
func (cdb *CredentialDB) GetUser(ctx context.Context, usr string, psw string) (uint8, error) {
	query := `SELECT password, permissions FROM users WHERE name = ?`

	row := cdb.db.QueryRowContext(ctx, query, usr)
	
	var dbPsw string
	var dbPerm uint8

	if err := row.Scan(&dbPsw, &dbPerm); err != nil {
		// If an error occurred check if there were no returned rows, meaning there's no entry with the given name.
		if err == sql.ErrNoRows {
			return 0, ErrUserNotExists
		}

		return 0, err
	}

	// Check password validity.
	if !CheckPasswordHash(psw, dbPsw) {
		return 0, ErrInvalidPassword
	}

	return dbPerm, nil
}