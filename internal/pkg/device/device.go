package device

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"

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
	Name     string
	Addr     net.Addr
	Services []*serv.Service
	Messages []*serv.Message
}

// NewDevice creates a new device by accepting a service definiton
func NewDevice(name string, address net.Addr, s []byte) (*Device, error) {
	srv, err := readService(s)
	if err != nil {
		return nil, err
	}
	var services []*serv.Service
	var messages []*serv.Message
	for _, def := range srv.Definitions {
		if def.Message == nil {
			services = append(services, def.Service)
		} else {
			messages = append(messages, def.Message)
		}
	}
	dev := &Device{
		ID:       uuid.New(),
		Name:     name,
		Addr:     address,
		Services: services,
		Messages: messages,
	}
	return dev, nil
}

// Call calls a service with a data
func (d *Device) Call(service string, query string) ([]byte, error) {
	connectionString := fmt.Sprintf("http://%v/%v?%v", d.Addr.String(), service, query)
	u, err := url.Parse(connectionString)
	if err != nil {
		return nil, err
	}

	if u.Scheme == "" || u.Host == "" || u.Path == "" {
		return nil, fmt.Errorf("Invalid URL")
	}
	log.Println("calling to " + connectionString)
	resp, err := http.Get(connectionString)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
