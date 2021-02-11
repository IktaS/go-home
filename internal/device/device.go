package device

import (
	"net"

	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/google/uuid"
)

func readService(input []byte) (*serv.Serv, error) {
	parser, err := serv.NewServParser()
	if err != nil {
		return nil, err
	}
	srv, err := parser.Parse(input)
	if err != nil {
		return nil, err
	}
	return srv, err
}

// Device defines a device id, and it's respective message and services
type Device struct {
	ID       uuid.UUID
	Addr     net.Addr
	Services []serv.Service
	Messages []serv.Message
}

// NewDevice creates a new device by accepting a service definiton
func NewDevice(address net.Addr, s []byte) (*Device, error) {
	srv, err := readService(s)
	if err != nil {
		return nil, err
	}
	var services []serv.Service
	var messages []serv.Message
	for _, def := range srv.Definitions {
		if def.Message == nil {
			services = append(services, *def.Service)
		} else {
			messages = append(messages, *def.Message)
		}
	}
	dev := &Device{
		ID:       uuid.New(),
		Addr:     address,
		Services: services,
		Messages: messages,
	}
	return dev, nil
}
