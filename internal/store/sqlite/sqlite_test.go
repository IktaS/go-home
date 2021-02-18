package sqlite

import (
	"fmt"
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
		setup    func(t *testing.T) (*Store, sqlmock.Sqlmock)
		teardown func(t *testing.T, s *Store)
		input    *device.Device
		expected func(sqlmock.Sqlmock, *device.Device)
		wantErr  bool
	}{
		{
			name: "Default Test",
			setup: func(t *testing.T) (*Store, sqlmock.Sqlmock) {
				db, mock, err := sqlmock.New()
				if err != nil {
					t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
				}
				s := &Store{
					FileName: "test_sqlite.db",
					DB:       db,
				}
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
			expected: func(mock sqlmock.Sqlmock, d *device.Device) {
				mock.ExpectBegin()

				mock.ExpectExec(
					regexp.QuoteMeta(fmt.Sprintf("INSERT OR IGNORE INTO devices(id, name, addr) VALUES(%v,%v,%v);", d.ID.String(), d.Name, d.Addr.String())),
				).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					regexp.QuoteMeta(fmt.Sprintf("INSERT OR IGNORE INTO messages(device_id, name) VALUES(%v,%v);", d.ID.String(), "TestMessage")),
				).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					regexp.QuoteMeta(fmt.Sprintf(`INSERT OR IGNORE INTO message_definition_fields(message_id, name, is_optional, is_required, is_scalar, value) 
									VALUES(%v,%v,%v,%v,%v,%v);`, 1, "TestString", 0, 0, 1, "string")),
				).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					regexp.QuoteMeta(fmt.Sprintf("INSERT OR IGNORE INTO service_response(is_scalar, value) VALUES(%v,%v);", 1, "string")),
				).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					regexp.QuoteMeta(fmt.Sprintf("INSERT OR IGNORE INTO services(device_id, name, response_id) VALUES(%v,%v,%v);", d.ID.String(), "TestService", 1)),
				).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec(
					regexp.QuoteMeta(fmt.Sprintf("INSERT OR IGNORE INTO service_request(service_id, is_scalar, value) VALUES(%v, %v,%v);", 1, 0, "TestMessage")),
				).WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, mock := tt.setup(t)
			tt.expected(mock, tt.input)
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
