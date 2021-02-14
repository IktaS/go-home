package store

import "github.com/IktaS/go-home/internal/device"

//Repo is an interface that defines what a repository should have
type Repo interface {
	Init() error
	Save(*device.Device) error
	Get(interface{}) (*device.Device, error)
	GetAll() ([]*device.Device, error)
	Delete(interface{}) error
}
