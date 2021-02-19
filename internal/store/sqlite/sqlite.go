package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net"
	"os"

	"github.com/IktaS/go-home/internal/device"
	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // import sqlite3 driver
)

func booltoI(b bool) int {
	if b {
		return 1
	}
	return 0
}

func intToBool(i int) bool {
	if i == 1 {
		return true
	}
	return false
}

func typeToDBModel(t *serv.Type) (int, string) {
	isScalar := (t.Reference == "")
	var value string
	if isScalar {
		value = t.Scalar.String()
	} else {
		value = t.Reference
	}
	return booltoI(isScalar), value
}

func dbModelToType(isScalar int, value string) *serv.Type {
	if isScalar == 1 {
		return &serv.Type{
			Scalar: serv.StringToScalar[value],
		}
	}
	return &serv.Type{
		Reference: value,
	}
}

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
		"response_id" INTEGER NOT NULL,
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
		"message_id" INTEGER NOT NULL,
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
	ctx := context.Background()
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	insertDeviceSQL := "INSERT OR IGNORE INTO devices(id, name, addr) VALUES(?,?,?);"
	_, err = tx.ExecContext(ctx, insertDeviceSQL, d.ID.String(), d.Name, d.Addr.String())
	if err != nil {
		tx.Rollback()
		return err
	}
	for _, m := range d.Messages {
		err := insertMessage(ctx, tx, d.ID, m)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, s := range d.Services {
		err := insertService(ctx, tx, d.ID, s)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func insertMessage(ctx context.Context, tx *sql.Tx, devID uuid.UUID, m *serv.Message) error {
	insertMessageSQL := "INSERT OR IGNORE INTO messages(device_id, name) VALUES(?,?);"
	row, err := tx.ExecContext(ctx, insertMessageSQL, devID.String(), m.Name)
	if err != nil {
		return err
	}
	messageID, err := row.LastInsertId()
	if err != nil {
		return err
	}
	for _, md := range m.Definitions {
		if md.Field != nil {
			err := insertMessageField(ctx, tx, messageID, md.Field)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func insertMessageField(ctx context.Context, tx *sql.Tx, mesID int64, f *serv.Field) error {
	isScalar, value := typeToDBModel(f.Type)
	insertMesDefSQL := `INSERT OR IGNORE INTO message_definition_fields(message_id, name, is_optional, is_required, is_scalar, value) 
						VALUES(?,?,?,?,?,?);`
	_, err := tx.ExecContext(ctx, insertMesDefSQL, mesID, f.Name, booltoI(f.Optional), booltoI(f.Required), isScalar, value)
	if err != nil {
		return err
	}
	return nil
}

func insertService(ctx context.Context, tx *sql.Tx, devID uuid.UUID, s *serv.Service) error {
	responseID, err := insertServiceResponse(ctx, tx, s.Response)
	if err != nil {
		return err
	}
	insertServiceSQL := "INSERT OR IGNORE INTO services(device_id, name, response_id) VALUES(?,?,?);"
	row, err := tx.ExecContext(ctx, insertServiceSQL, devID.String(), s.Name, responseID)
	if err != nil {
		return err
	}
	serviceID, err := row.LastInsertId()
	if err != nil {
		return err
	}
	for _, r := range s.Request {
		err := insertServiceRequest(ctx, tx, serviceID, r)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertServiceResponse(ctx context.Context, tx *sql.Tx, t *serv.Type) (int64, error) {
	isScalar, value := typeToDBModel(t)
	insertServiceResponseSQL := "INSERT OR IGNORE INTO service_response(is_scalar, value) VALUES(?,?);"
	row, err := tx.ExecContext(ctx, insertServiceResponseSQL, isScalar, value)
	if err != nil {
		return -1, err
	}
	id, err := row.LastInsertId()
	if err != nil {
		return -1, err
	}
	return id, nil
}

func insertServiceRequest(ctx context.Context, tx *sql.Tx, id int64, t *serv.Type) error {
	isScalar, value := typeToDBModel(t)
	insertServiceRequestSQL := "INSERT OR IGNORE INTO service_request(service_id, is_scalar, value) VALUES(?,?,?);"
	_, err := tx.ExecContext(ctx, insertServiceRequestSQL, id, isScalar, value)
	if err != nil {
		return err
	}
	return nil
}

// Get defines getting a device.Device
func (p *Store) Get(id interface{}) (*device.Device, error) {
	id = id.(string)
	deviceQuerySQL := "SELECT * FROM devices WHERE id = ?"
	deviceRow := p.DB.QueryRow(deviceQuerySQL, id)
	var uuID string
	var name string
	var addr string
	err := deviceRow.Scan(&uuID, &name, &addr)
	if err != nil {
		return nil, err
	}
	dev, err := dbDeviceToDevice(p.DB, uuID, name, addr)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

func dbDeviceToDevice(db *sql.DB, id string, name string, addr string) (*device.Device, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	dev := &device.Device{
		ID:   uid,
		Name: name,
		Addr: &net.IPAddr{
			IP:   net.ParseIP(addr),
			Zone: "",
		},
	}
	messageQuerySQL := "SELECT * FROM messages WHERE device_id = ?"
	messageRows, err := db.Query(messageQuerySQL, id)
	defer messageRows.Close()
	if err != nil {
		return nil, err
	}
	dev.Messages, err = messageRowsToMessages(db, messageRows)
	if err != nil {
		return nil, err
	}

	serviceQuerySQL := "SELECT * FROM services WHERE device_id = ?"
	serviceRows, err := db.Query(serviceQuerySQL, id)
	defer serviceRows.Close()
	if err != nil {
		return nil, err
	}
	dev.Services, err = serviceRowsToServices(db, serviceRows)
	if err != nil {
		return nil, err
	}
	return dev, nil
}

func messageFieldRowsToMessageDefinition(db *sql.DB, mesID int) ([]*serv.MessageDefinition, error) {
	messageFieldSQL := "SELECT * FROM message_definition_fields WHERE message_id = ?"
	messageFieldRows, err := db.Query(messageFieldSQL, mesID)
	defer messageFieldRows.Close()
	if err != nil {
		return nil, err
	}
	var mesDef []*serv.MessageDefinition
	for messageFieldRows.Next() {
		var id int
		var messageID int
		var name string
		var isOptional int
		var isRequired int
		var isScalar int
		var value string
		err = messageFieldRows.Scan(&id, &messageID, &name, &isOptional, &isRequired, &isScalar, &value)
		if err != nil {
			return nil, err
		}
		mesDef = append(mesDef, &serv.MessageDefinition{
			Field: &serv.Field{
				Optional: intToBool(isOptional),
				Required: intToBool(isRequired),
				Type:     dbModelToType(isScalar, value),
				Name:     name,
			},
		})
	}
	return mesDef, nil
}

func getMessageDefinition(db *sql.DB, mesID int) ([]*serv.MessageDefinition, error) {
	var messageDefinitions []*serv.MessageDefinition

	fields, err := messageFieldRowsToMessageDefinition(db, mesID)
	if err != nil {
		return nil, err
	}
	messageDefinitions = append(messageDefinitions, fields...)

	//in case future message definition type is added, add code here

	return messageDefinitions, nil
}

func messageRowsToMessages(db *sql.DB, rows *sql.Rows) ([]*serv.Message, error) {
	var messages []*serv.Message
	for rows.Next() {
		var id int
		var deviceID string
		var name string
		err := rows.Scan(&id, &deviceID, &name)
		if err != nil {
			return nil, err
		}
		messageDefinitions, err := getMessageDefinition(db, id)
		if err != nil {
			return nil, err
		}

		message := &serv.Message{
			Name:        name,
			Definitions: messageDefinitions,
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func getServiceResponse(db *sql.DB, id int) (*serv.Type, error) {
	serviceResponseSQL := "SELECT * FROM service_response WHERE id = ?"
	serviceResponseRow := db.QueryRow(serviceResponseSQL, id)
	var isScalar int
	var value string
	err := serviceResponseRow.Scan(&id, &isScalar, &value)
	if err != nil {
		return nil, err
	}
	return dbModelToType(isScalar, value), nil
}

func getServiceRequest(db *sql.DB, serviceID int) ([]*serv.Type, error) {
	serviceRequestSQL := "SELECT * FROM service_request WHERE service_id = ?"
	serviceRequestRows, err := db.Query(serviceRequestSQL, serviceID)
	if err != nil {
		return nil, err
	}
	defer serviceRequestRows.Close()
	var requests []*serv.Type
	for serviceRequestRows.Next() {
		var serviceID int
		var isScalar int
		var value string
		err = serviceRequestRows.Scan(&serviceID, &isScalar, &value)
		if err != nil {
			return nil, err
		}
		requests = append(requests, dbModelToType(isScalar, value))
	}
	return requests, nil
}

func serviceRowsToServices(db *sql.DB, rows *sql.Rows) ([]*serv.Service, error) {
	var services []*serv.Service
	for rows.Next() {
		var id int
		var deviceID string
		var name string
		var responseID int
		err := rows.Scan(&id, &deviceID, &name, &responseID)
		if err != nil {
			return nil, err
		}
		response, err := getServiceResponse(db, responseID)
		if err != nil {
			return nil, err
		}
		requests, err := getServiceRequest(db, id)
		if err != nil {
			return nil, err
		}
		services = append(services, &serv.Service{
			Name:     name,
			Request:  requests,
			Response: response,
		})
	}
	return services, nil
}

// GetAll gets all device
func (p *Store) GetAll() ([]*device.Device, error) {
	return nil, errors.New("Not Implemented")
}

// Delete defines getting a device.Device
func (p *Store) Delete(id interface{}) error {
	return errors.New("Not Implemented")
}
