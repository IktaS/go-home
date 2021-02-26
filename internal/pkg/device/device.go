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

//ToJSON returns a json string that represent the device
func (d *Device) ToJSON() string {
	ret := fmt.Sprintf("{\"id\":\"%v\",\"name\":\"%v\",\"services\":%v,\"messages\":%v}",
		d.ID.String(),
		d.Name,
		serviceArrayToJSON(d.Services),
		messageArrayToJSON(d.Messages),
	)
	return ret
}

func messageArrayToJSON(messages []*serv.Message) string {
	ret := "["
	notfirst := false
	for _, m := range messages {
		if notfirst {
			ret += ","
		}
		notfirst = true
		ret += messageToJSON(m)
	}
	ret += "]"
	return ret
}

func messageToJSON(m *serv.Message) string {
	ret := fmt.Sprintf("{\"name\":\"%v\",\"definitions\":%v}",
		m.Name,
		mesDefinitionArrayToJSON(m.Definitions),
	)
	return ret
}

func mesDefinitionArrayToJSON(mesDefs []*serv.MessageDefinition) string {
	ret := "["
	notfirst := false
	for _, m := range mesDefs {
		if notfirst {
			ret += ","
		}
		notfirst = true
		ret += mesDefToJSON(m)
	}
	ret += "]"
	return ret
}

func mesDefToJSON(m *serv.MessageDefinition) string {
	if m.Field != nil {
		f := m.Field
		isOptional := f.Optional && !f.Required
		return fmt.Sprintf("{\"name\":\"%v\",\"isOptional\":\"%v\",\"value\":%v}",
			f.Name,
			isOptional,
			typeToJSON(f.Type),
		)
	}
	return "\"None\""
}

func serviceArrayToJSON(services []*serv.Service) string {
	ret := "["
	notfirst := false
	for _, s := range services {
		if notfirst {
			ret += ","
		}
		notfirst = true
		ret += serviceToJSON(s)
	}
	ret += "]"
	return ret
}

func serviceToJSON(s *serv.Service) string {
	ret := fmt.Sprintf("{\"name\":\"%v\",\"response\":%v,\"request\":%v}",
		s.Name,
		typeToJSON(s.Response),
		typeArrayToJSON(s.Request),
	)
	return ret
}

func typeArrayToJSON(types []*serv.Type) string {
	ret := "["
	notfirst := false
	for _, t := range types {
		if notfirst {
			ret += ","
		}
		notfirst = true
		ret += typeToJSON(t)
	}
	ret += "]"
	return ret
}

func typeToJSON(t *serv.Type) string {
	if t.Reference == "" {
		return fmt.Sprintf("{\"isScalar\":\"true\",\"value\":\"%v\"}", t.Scalar.String())
	}
	return fmt.Sprintf("{\"isScalar\":\"false\",\"value\":\"%v\"}", t.Reference)
}
