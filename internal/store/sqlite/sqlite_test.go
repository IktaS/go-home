package sqlite

import (
	"os"
	"testing"

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
