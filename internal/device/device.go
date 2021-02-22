package device

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

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
func (d *Device) Call(service string, data map[string]interface{}) ([]byte, error) {
	jsondata, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	connectionString := fmt.Sprintf("%v/%v", d.Addr.String(), service)
	log.Println("calling to " + connectionString)
	req, err := http.NewRequest("POST", connectionString, bytes.NewBuffer(jsondata))
	client := &http.Client{}
	resp, err := client.Do(req)
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
