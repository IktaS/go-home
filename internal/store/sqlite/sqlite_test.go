package sqlite

import (
	"net"
	"os"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/IktaS/go-home/internal/device"
	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestNewSQLiteStore(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, filename string)
		filename string
		teardown func(t *testing.T, filename string)
		wantErr  bool
	}{
		{
			name:     "Normal run",
			setup:    func(t *testing.T, filename string) {},
			filename: "sqlite-db.db",
			teardown: func(t *testing.T, filename string) {
				err := os.Remove(filename)
				if err != nil {
					t.Fatal(err)
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t, tt.filename)
			_, err := NewSQLiteStore(tt.filename)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.teardown(t, tt.filename)
		})
	}
}

func TestStore_Save(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, d *device.Device) (*Store, sqlmock.Sqlmock)
		teardown func(t *testing.T, s *Store)
		input    *device.Device
		wantErr  bool
	}{
		{
			name: "Default Test",
			setup: func(t *testing.T, d *device.Device) (*Store, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				s := &Store{
					FileName: "test_sqlite.db",
					DB:       db,
				}

				mock.ExpectBegin()

				mock.ExpectExec(
					"INSERT OR IGNORE INTO devices",
				).WithArgs(d.ID.String(), d.Name, d.Addr.String()).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					"INSERT OR IGNORE INTO messages",
				).WithArgs(d.ID.String(), "TestMessage").WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					"INSERT OR IGNORE INTO message_definition_fields",
				).WithArgs(1, "TestString", 0, 0, 1, "string").WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					"INSERT OR IGNORE INTO service_response",
				).WithArgs(1, "string").WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					"INSERT OR IGNORE INTO services",
				).WithArgs(d.ID.String(), "TestService", 1).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					"INSERT OR IGNORE INTO service_request",
				).WithArgs(1, 0, "TestMessage").WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()

				return s, mock
			},
			teardown: func(t *testing.T, s *Store) {
				s.DB.Close()
			},
			input: &device.Device{
				ID:   uuid.New(),
				Name: "test-device",
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 80,
				},
				Services: []*serv.Service{
					{
						Name: "TestService",
						Request: []*serv.Type{
							{
								Reference: "TestMessage",
							},
						},
						Response: &serv.Type{
							Scalar: serv.String,
						},
					},
				},
				Messages: []*serv.Message{
					{
						Name: "TestMessage",
						Definitions: []*serv.MessageDefinition{
							{
								Field: &serv.Field{
									Name: "TestString",
									Type: &serv.Type{
										Scalar: serv.String,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, mock := tt.setup(t, tt.input)
			err := p.Save(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			err = mock.ExpectationsWereMet()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.teardown(t, p)
		})
	}
}

func TestStore_Get(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, id string) (*Store, sqlmock.Sqlmock)
		teardown func(t *testing.T, s *Store)
		input    string
		expected *device.Device
		wantErr  bool
	}{
		{
			name: "Default Test",
			setup: func(t *testing.T, deviceID string) (*Store, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				s := &Store{
					FileName: "test_sqlite.db",
					DB:       db,
				}

				// setup database filling
				//device filling
				deviceRows := sqlmock.NewRows([]string{"id", "name", "addr"}).
					AddRow(deviceID, "test-device", "127.0.0.1:80")
				mock.ExpectQuery(
					regexp.QuoteMeta("SELECT * FROM devices WHERE id"),
				).WithArgs(deviceID).WillReturnRows(deviceRows)

				//messages filling
				messageID := 1
				messageRows := sqlmock.NewRows([]string{"id", "device_id", "name"}).
					AddRow(messageID, deviceID, "TestMessage")
				mock.ExpectQuery(
					regexp.QuoteMeta("SELECT * FROM messages WHERE device_id"),
				).WithArgs(deviceID).WillReturnRows(messageRows)
				mesDefRows := sqlmock.NewRows([]string{"id", "message_id", "name", "is_optional", "is_required", "is_scalar", "value"}).
					AddRow(1, messageID, "TestString", 0, 0, 1, "string")
				mock.ExpectQuery(
					regexp.QuoteMeta("SELECT * FROM message_definition_fields WHERE message_id"),
				).WithArgs(messageID).WillReturnRows(mesDefRows)

				//service filling
				serviceID := 1
				responseID := 1

				serviceRows := sqlmock.NewRows([]string{"id", "device_id", "name", "response_id"}).
					AddRow(serviceID, deviceID, "TestService", responseID)
				mock.ExpectQuery(
					regexp.QuoteMeta("SELECT * FROM services WHERE device_id"),
				).WithArgs(deviceID).WillReturnRows(serviceRows)

				responseRows := sqlmock.NewRows([]string{"id", "is_scalar", "value"}).
					AddRow(responseID, 1, "string")
				mock.ExpectQuery(
					regexp.QuoteMeta("SELECT * FROM service_response WHERE id"),
				).WithArgs(responseID).WillReturnRows(responseRows)

				requestRows := sqlmock.NewRows([]string{"id", "service_id", "is_scalar", "value"}).
					AddRow(1, serviceID, 0, "TestMessage")
				mock.ExpectQuery(
					regexp.QuoteMeta("SELECT * FROM service_request WHERE service_id"),
				).WithArgs(serviceID).WillReturnRows(requestRows)
				return s, mock
			},
			teardown: func(t *testing.T, s *Store) {
				s.DB.Close()
			},
			expected: &device.Device{
				ID:   uuid.New(),
				Name: "test-device",
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 80,
				},
				Services: []*serv.Service{
					{
						Name: "TestService",
						Request: []*serv.Type{
							{
								Reference: "TestMessage",
							},
						},
						Response: &serv.Type{
							Scalar: serv.String,
						},
					},
				},
				Messages: []*serv.Message{
					{
						Name: "TestMessage",
						Definitions: []*serv.MessageDefinition{
							{
								Field: &serv.Field{
									Name: "TestString",
									Type: &serv.Type{
										Scalar: serv.String,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, mock := tt.setup(t, tt.expected.ID.String())
			ret, err := p.Get(tt.expected.ID.String())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			err = mock.ExpectationsWereMet()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expected, ret)
			tt.teardown(t, p)
		})
	}
}

func TestStore_Delete(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T) (*Store, sqlmock.Sqlmock)
		teardown func(t *testing.T, s *Store)
		input    string
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, mock := tt.setup(t)
			err := p.Delete(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			err = mock.ExpectationsWereMet()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			tt.teardown(t, p)
		})
	}
}
