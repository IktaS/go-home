package sqlite

import (
	"database/sql"
	"errors"
	"log"
	"os"

	"github.com/IktaS/go-home/internal/device"
	_ "github.com/lib/pq"
)

//Store defines what the Postgre SQL Store needs
type Store struct {
	FileName string
}

// NewPostgreSQLStore makes a new PostgreSQL Store
func NewSQLiteStore(dsn string) (*Store, error) {
	p := &Store{FileName: dsn}
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
			log.Fatal(err.Error())
		}
		file.Close()
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		log.Fatal(err.Error())
	}
	db, err := sql.Open("sqlite3", "./"+p.FileName)
	defer db.Close()

	//TODO: Code to create tables

	return err
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
