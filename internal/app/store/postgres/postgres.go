package postgres

import (
	"database/sql"
	"errors"

	"github.com/IktaS/go-home/internal/pkg/device"
	_ "github.com/lib/pq"
)

//Store defines what the Postgre SQL Store needs
type Store struct {
	DSN string
	DB  *sql.DB
}

// NewPostgreSQLStore makes a new PostgreSQL Store
func NewPostgreSQLStore(dsn string) (*Store, error) {
	p := &Store{DSN: dsn}
	err := p.Init()
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Init initialize a postgreSQL
func (p *Store) Init() error {
	db, err := sql.Open("postgres", p.DSN)
	if err != nil {
		return err
	}
	p.DB = db
	return nil
}

// Save saves a device to the postgreSQL store
func (p *Store) Save(d *device.Device) error {
	return errors.New("Not Implemented")
}

// Get defines getting a device.Device
func (p *Store) Get(id interface{}) (*device.Device, error) {
	return nil, errors.New("Not Implemented")
}

// GetAll gets all device
func (p *Store) GetAll() ([]*device.Device, error) {
	return nil, errors.New("Not Implemented")
}

// Delete defines getting a device.Device
func (p *Store) Delete(id interface{}) error {
	return errors.New("Not Implemented")
}
