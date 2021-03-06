package device

import (
	"net"
	"testing"

	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/stretchr/testify/assert"
)

func TestNewDevice(t *testing.T) {
	tests := []struct {
		Name     string
		DevName  string
		input    []byte
		expected *Device
		WantErr  bool
	}{
		{
			Name:    "Default test",
			DevName: "Device1",
			input: []byte(`
				message TestMessage{string TestString;};def inbound TestService(TestMessage):string;
			`),
			expected: &Device{
				Name: "Device1",
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 80,
				},
				Services: []*serv.Service{
					{
						Name:     "TestService",
						Inbound:  true,
						Outbound: false,
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
		},
		{
			Name:    "One Message",
			DevName: "Device1",
			input: []byte(`
				message TestMessage{string TestString;};
			`),
			expected: &Device{
				Name: "Device1",
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 80,
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
		},
		{
			Name:    "One Service",
			DevName: "Device1",
			input: []byte(`
				def outbound click();
			`),
			expected: &Device{
				Name: "Device1",
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 80,
				},
				Services: []*serv.Service{
					{
						Name:     "click",
						Inbound:  false,
						Outbound: true,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			addr := &net.TCPAddr{
				IP:   net.IPv4(127, 0, 0, 1),
				Port: 80,
			}
			dev, err := NewDevice(tt.DevName, addr, tt.input)
			if tt.WantErr {
				assert.Error(t, err)
			}
			tt.expected.ID = dev.ID
			assert.Equal(t, tt.expected, dev)
		})
	}
}
