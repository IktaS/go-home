package postgres

import (
	"database/sql"
	"net"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/IktaS/go-home/internal/device"
	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgreSQLStore_Save(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) (*Store, sqlmock.Sqlmock, *sql.DB)
		teardown func(*testing.T, *sql.DB)
		input    *device.Device
		wantErr  bool
	}{
		{
			name: "Default Test",
			setup: func(t *testing.T) (*Store, sqlmock.Sqlmock, *sql.DB) {
				db, mock, err := sqlmock.New() // mock sql.DB
				if err != nil {
					t.Fatal(err)
				}
				gdb, err := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{}) // open gorm db
				if err != nil {
					t.Fatal(err)
				}

				//TODO: add expectations

				r := &Store{
					DB: gdb,
				}
				return r, mock, db
			},
			teardown: func(t *testing.T, db *sql.DB) {
				db.Close()
			},
			input: &device.Device{
				ID:   uuid.New(),
				Name: "Test Device",
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
			p, mock, db := tt.setup(t)
			err := p.Save(tt.input)
			if err != nil {
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
			err = mock.ExpectationsWereMet()
			if err != nil {
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}
			tt.teardown(t, db)
		})
	}
}

func TestPostgreSQLStore_Get(t *testing.T) {
	type fields struct {
		DSN string
		DB  *gorm.DB
	}
	type args struct {
		id interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *device.Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Store{
				DSN: tt.fields.DSN,
				DB:  tt.fields.DB,
			}
			got, err := p.Get(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgreSQLStore.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgreSQLStore.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgreSQLStore_GetAll(t *testing.T) {
	type fields struct {
		DSN string
		DB  *gorm.DB
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*device.Device
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Store{
				DSN: tt.fields.DSN,
				DB:  tt.fields.DB,
			}
			got, err := p.GetAll()
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgreSQLStore.GetAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostgreSQLStore.GetAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPostgreSQLStore_Delete(t *testing.T) {
	type fields struct {
		DSN string
		DB  *gorm.DB
	}
	type args struct {
		id interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Store{
				DSN: tt.fields.DSN,
				DB:  tt.fields.DB,
			}
			if err := p.Delete(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("PostgreSQLStore.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
