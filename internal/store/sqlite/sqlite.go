package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/IktaS/go-home/internal/device"
	_ "github.com/mattn/go-sqlite3" // import sqlite3 driver
)

//Store defines what the Postgre SQL Store needs
type Store struct {
	FileName string
}

// NewSQLiteStore makes a new SQLite Store
func NewSQLiteStore(filename string) (*Store, error) {
	p := &Store{FileName: filename}
	err := p.Init()
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Init initialize a postgreSQL
func (p *Store) Init() error {
	if _, err := os.Stat(p.FileName); err == nil {
		// sqlite_database.db exists
		log.Println("Database exist, skipped making database file")
	} else if os.IsNotExist(err) {
		// sqlite_database.db does *not* exist
		file, err := os.Create(p.FileName)
		if err != nil {
			return err
		}
		file.Close()
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return err
	}
	db, err := sql.Open("sqlite3", "./"+p.FileName)
	defer db.Close()

	//TODO: Code to create tables

	//enable foreign key
	statement, err := db.Prepare(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		return err
	}
	statement.Exec()

	// Create Devices Table
	createDevicesTableSQL := `CREATE TABLE IF NOT EXISTS devices(
		"id" TEXT NOT NULL PRIMARY KEY,
		"name" TEXT,
		"addr" TEXT
	);`

	statement, err = db.Prepare(createDevicesTableSQL)
	if err != nil {
		return err
	}
	statement.Exec()

	// Create ServiceResponse Table
	createServiceResponseTableSQL := `CREATE TABLE IF NOT EXISTS service_response(
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"is_scalar" INTEGER,
		"value" TEXT
	);`

	statement, err = db.Prepare(createServiceResponseTableSQL)
	if err != nil {
		return err
	}
	statement.Exec()

	// Create Services Table
	createServicesTableSQL := `CREATE TABLE IF NOT EXISTS services(
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"device_id" TEXT NOT NULL,
		"name" TEXT,
		"response_id" int NOT NULL,
		FOREIGN KEY (device_id) REFERENCES devices (id),
		FOREIGN KEY (response_id) REFERENCES service_response (id)
	);`

	statement, err = db.Prepare(createServicesTableSQL)
	if err != nil {
		return err
	}
	statement.Exec()

	// Create ServiceRequest Table
	createServiceRequestTableSQL := `CREATE TABLE IF NOT EXISTS service_request(
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"service_id" INTEGER NOT NULL,
		"is_scalar" INTEGER,
		"value" TEXT
	);`

	statement, err = db.Prepare(createServiceRequestTableSQL)
	if err != nil {
		return err
	}
	statement.Exec()

	// Create Messages Table
	createMessagesTableSQL := `CREATE TABLE IF NOT EXISTS messages(
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"device_id" TEXT NOT NULL,
		"name" TEXT,
		FOREIGN KEY (device_id) REFERENCES devices (id)
	);`

	statement, err = db.Prepare(createMessagesTableSQL)
	if err != nil {
		return err
	}
	statement.Exec()

	// Create MessageDefinitions Table
	createMessageDefinitionsTableSQL := `CREATE TABLE IF NOT EXISTS message_definitions(
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"message_id" TEXT NOT NULL,
		"name" TEXT,
		"is_optional" INTEGER,
		"is_required" INTEGER,
		FOREIGN KEY (message_id) REFERENCES messages (id)
	);`

	statement, err = db.Prepare(createMessageDefinitionsTableSQL)
	if err != nil {
		return err
	}
	statement.Exec()

	// Create MessageDefinitionFields Table
	createMessageDefinitionFieldsTableSQL := `CREATE TABLE IF NOT EXISTS message_definition_fields(
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"message_definition_id" TEXT NOT NULL,
		"is_scalar" INTEGER,
		"value" TEXT,
		FOREIGN KEY (message_definition_id) REFERENCES message_definitions (id)
	);`

	statement, err = db.Prepare(createMessageDefinitionFieldsTableSQL)
	if err != nil {
		return err
	}
	statement.Exec()

	return nil
}

// Save saves a device to the postgreSQL store
func (p *Store) Save(d *device.Device) error {
	db, err := sql.Open("sqlite3", "./"+p.FileName)
	if err != nil {
		return err
	}
	defer db.Close()
	return errors.New("Not Implemented")
}

// Get defines getting a device.Device
func (p *Store) Get(id interface{}) (*device.Device, error) {
	db, err := sql.Open("sqlite3", "./"+p.FileName)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return nil, errors.New("Not Implemented")
}

// GetAll gets all device
func (p *Store) GetAll() ([]*device.Device, error) {
	db, err := sql.Open("sqlite3", "./"+p.FileName)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return nil, errors.New("Not Implemented")
}

// Delete defines getting a device.Device
func (p *Store) Delete(id interface{}) error {
	db, err := sql.Open("sqlite3", "./"+p.FileName)
	if err != nil {
		return err
	}
	defer db.Close()
	return errors.New("Not Implemented")
}
