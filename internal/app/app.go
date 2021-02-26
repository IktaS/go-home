package app

import (
	"github.com/IktaS/go-home/internal/app/store"
)

// App devices what the app have
type App struct {
	Devices store.Repo
}

// NewApp creates a new app
func NewApp(r store.Repo) *App {
	return &App{
		Devices: r,
	}
}
