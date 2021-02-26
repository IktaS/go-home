package postgres

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/IktaS/go-home/internal/pkg/device"
	"github.com/stretchr/testify/assert"
)

func TestPostgreSQLStore_Save(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) (*Store, sqlmock.Sqlmock, *sql.DB)
		teardown func(*testing.T, *sql.DB)
		input    *device.Device
		expect   func(sqlmock.Sqlmock)
		wantErr  bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, mock, db := tt.setup(t)

			err := p.Save(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.expect(mock)
			err = mock.ExpectationsWereMet()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tt.teardown(t, db)
		})
	}
}
