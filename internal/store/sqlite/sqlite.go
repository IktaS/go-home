package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/IktaS/go-home/internal/device"
	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // import sqlite3 driver
)

//Store defines what the Postgre SQL Store needs
type Store struct {
	FileName string
	DB       *sql.DB
}

// NewSQLiteStore makes a new SQLite Store
func NewSQLiteStore(filename string) (*Store, error) {
	p := &Store{}
	err := p.Init(filename)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Init initialize a SQLite
func (p *Store) Init(config interface{}) error {
	filename := config.(string)
	if _, err := os.Stat(filename); err == nil {
		// database exists
		log.Println("Database exist, skipped making database file")
	} else if os.IsNotExist(err) {
		// database does not exist
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		file.Close()
	} else {
		// Schrodinger: file may or may not exist. See err for details.

		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
		return err
	}
	db, err := sql.Open("sqlite3", "./"+filename)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}

	//enable foreign key
	statement, err := db.Prepare(`PRAGMA foreign_keys = ON;`)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

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
	_, err = statement.Exec()
	if err != nil {
		return err
	}

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
	_, err = statement.Exec()
	if err != nil {
		return err
	}

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
	_, err = statement.Exec()
	if err != nil {
		return err
	}

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
	_, err = statement.Exec()
	if err != nil {
		return err
	}

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
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	// Create MessageDefinitionFields Table
	createMessageDefinitionFieldsTableSQL := `CREATE TABLE IF NOT EXISTS message_definition_fields(
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"message_id" TEXT NOT NULL,
		"name" TEXT,
		"is_optional" TEXT,
		"is_required" TEXT,
		"is_scalar" INTEGER,
		"value" TEXT,
		FOREIGN KEY (message_id) REFERENCES messages(id)
	);`

	statement, err = db.Prepare(createMessageDefinitionFieldsTableSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}

	return nil
}

// Save saves a device to the SQLite store
func (p *Store) Save(d *device.Device) error {
	db := p.DB

	insertDeviceSQL := fmt.Sprintf("INSERT OR IGNORE INTO devices(id, name, addr) VALUES(%v,%v,%v);", d.ID.String(), d.Name, d.Addr.String())
	statement, err := db.Prepare(insertDeviceSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	for _, m := range d.Messages {
		err := insertMessage(db, d.ID, m)
		if err != nil {
			return err
		}
	}
	for _, s := range d.Services {
		err := insertService(db, d.ID, s)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertMessage(db *sql.DB, devID uuid.UUID, m *serv.Message) error {
	insertMessageSQL := fmt.Sprintf("INSERT OR IGNORE INTO messages(device_id, name) VALUES(%v,%v);", devID.String(), m.Name)
	statement, err := db.Prepare(insertMessageSQL)
	if err != nil {
		return err
	}
	row, err := statement.Exec()
	if err != nil {
		return err
	}
	messageID, err := row.LastInsertId()
	if err != nil {
		return err
	}
	for _, md := range m.Definitions {
		if md.Field != nil {
			err := insertMessageField(db, messageID, md.Field)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func insertMessageField(db *sql.DB, mesID int64, f *serv.Field) error {
	isScalar, value := typeToDBModel(f.Type)
	insertMesDefSQL := fmt.Sprintf(`INSERT OR IGNORE INTO message_definition_fields(message_id, name, is_optional, is_required, is_scalar, value) 
									VALUES(%v,%v,%v,%v,%v,%v);`, mesID, f.Name, f.Optional, f.Required, isScalar, value)
	statement, err := db.Prepare(insertMesDefSQL)
	if err != nil {
		return err
	}
	row, err := statement.Exec()
	if err != nil {
		return err
	}
	_, err = row.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func insertService(db *sql.DB, devID uuid.UUID, s *serv.Service) error {
	responseID, err := insertServiceResponse(db, s.Response)
	if err != nil {
		return err
	}
	insertServiceSQL := fmt.Sprintf("INSERT OR IGNORE INTO services(device_id, name, response_id) VALUES(%v,%v,%v);", devID.String(), s.Name, responseID)
	statement, err := db.Prepare(insertServiceSQL)
	if err != nil {
		return err
	}
	row, err := statement.Exec()
	if err != nil {
		return err
	}
	serviceID, err := row.LastInsertId()
	if err != nil {
		return err
	}
	for _, r := range s.Request {
		err := insertServiceRequest(db, serviceID, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func typeToDBModel(t *serv.Type) (bool, string) {
	isScalar := !(t.Reference == "")
	var value string
	if isScalar {
		value = t.Scalar.String()
	} else {
		value = t.Reference
	}
	return isScalar, value
}

func insertServiceResponse(db *sql.DB, t *serv.Type) (int64, error) {
	isScalar, value := typeToDBModel(t)
	insertServiceResponseSQL := fmt.Sprintf("INSERT OR IGNORE INTO service_response(is_scalar, value) VALUES(%v,%v);", isScalar, value)
	statement, err := db.Prepare(insertServiceResponseSQL)
	if err != nil {
		return -1, err
	}
	row, err := statement.Exec()
	if err != nil {
		return -1, err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func insertServiceRequest(db *sql.DB, id int64, t *serv.Type) error {
	isScalar, value := typeToDBModel(t)
	insertServiceRequestSQL := fmt.Sprintf("INSERT OR IGNORE INTO service_request(service_id, is_scalar, value) VALUES(%v, %v,%v);", id, isScalar, value)
	statement, err := db.Prepare(insertServiceRequestSQL)
	if err != nil {
		return err
	}
	_, err = statement.Exec()
	if err != nil {
		return err
	}
	return nil
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
