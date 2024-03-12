package whatsapp

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	Client *sql.DB
}

func CheckAndCreatePath(path, file string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			log.Fatal(errDir)
		}
	}
}

func StoreConnect(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return &Store{Client: db}, nil
}

func (s *Store) Create(
	id, name, phoneNumber, apikey, state, connection, containerName string,
	createdAt time.Time,
) error {
	sqlStmt := `
		CREATE TABLE IF NOT EXISTS instances (
			id TEXT,
			name TEXT,
			phoneNumber TEXT,
			state TEXT,
			apikey TEXT,
			connection TEXT,
			containerName TEXT,
			createdAt TIMESTAMP
		);
	`
	_, err := s.Client.Exec(sqlStmt)
	if err != nil {
		return err
	}

	stmt, err := s.Client.Prepare(`INSERT INTO instances(
		id,
		name,
		phoneNumber,
		state,
		apikey,
		connection,
		containerName,
		createdAt
	)
	VALUES(?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	formatDate := createdAt.Format("2006-01-02 15:04:05")

	_, err = stmt.Exec(id, name, phoneNumber, apikey, state, connection, containerName, formatDate)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Read(query string) (*Instance, error) {
	stmt, err := s.Client.Prepare(`
		SELECT * FROM instances 
		WHERE phoneNumber = ?
		OR id = ?
		OR name = ?
	`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(query, query, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var id, name, phoneNumber, state, apikey, connection, containerName string
		var createdAt time.Time
		err := rows.Scan(&id, &name, &phoneNumber, &apikey, &state, &connection, &containerName, &createdAt)
		if err != nil {
			return nil, err
		}

		instance := NewInstance(id, name, "", phoneNumber, apikey, StateEnum(state), nil, nil, containerName)
		instance.Connection = ConnectionStatusEnum(connection)
		instance.CreatedAt = createdAt

		return instance, nil
	}

	return nil, nil
}

func (s *Store) ReadAll() ([]*Instance, error) {
	stmt, err := s.Client.Prepare("SELECT * FROM instances")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instances := make([]*Instance, 0)
	for rows.Next() {
		var id, name, phoneNumber, state, apikey, connection, containerName string
		var createdAt time.Time
		err := rows.Scan(&id, &name, &phoneNumber, &apikey, &state, &connection, &containerName, &createdAt)
		if err != nil {
			return nil, err
		}

		instance := Instance{
			ID:         id,
			Name:       name,
			Apikey:     &apikey,
			WhatsApp:   &WhatsApp{Number: phoneNumber},
			Connection: ConnectionStatusEnum(connection),
			State:      StateEnum(state),
			CreatedAt:  createdAt,
		}

		instances = append(instances, &instance)
	}

	return instances, nil
}

func (s *Store) Update(id string, update *Instance) error {
	var build strings.Builder
	var args []interface{}

	build.WriteString("UPDATE instances SET ")

	first := true
	if update.State != "" {
		if !first {
			build.WriteString(", ")
		}
		build.WriteString("state = ?")
		args = append(args, update.State)
		first = false
	}
	if update.Connection != "" {
		if !first {
			build.WriteString(", ")
		}
		build.WriteString("connection = ?")
		args = append(args, update.Connection)
		first = false
	}
	if update.WhatsApp != nil && update.WhatsApp.Number != "" {
		if !first {
			build.WriteString(", ")
		}
		build.WriteString("phoneNumber = ?")
		args = append(args, update.WhatsApp.Number)
		first = false
	}
	if update.Name != "" {
		if !first {
			build.WriteString(", ")
		}
		build.WriteString("name = ?")
		args = append(args, update.Name)
		first = false
	}

	if id != "" {
		build.WriteString(" WHERE id = ?")
		args = append(args, id)
	}

	stmt, err := s.Client.Prepare(build.String())
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) Delete(id string) error {
	stmt, err := s.Client.Prepare("DELETE FROM instances WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}

	return nil
}
