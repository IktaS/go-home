package device

import (
	"net"
	"testing"

	"github.com/IktaS/go-serv/pkg/serv"
	"github.com/google/uuid"
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
				message TestMessage{string TestString;};service TestService(TestMessage):string;
			`),
			expected: &Device{
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
		},
		{
			Name:    "One Message",
			DevName: "Device1",
			input: []byte(`
				message TestMessage{string TestString;};
			`),
			expected: &Device{
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

func TestDevice_ToJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    *Device
		expected string
	}{
		// TODO: Add test cases.
		{
			name: "Default Test",
			input: &Device{
				ID:   uuid.MustParse("1299136e-c15f-4815-b667-492f72476818"),
				Name: "DevTest1",
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
			expected: `{"id":"1299136e-c15f-4815-b667-492f72476818","name":"DevTest1","services":[{"name":"TestService","response":{"isScalar":"true","value":"string"},"request":[{"isScalar":"false","value":"TestMessage"}]}],"messages":[{"name":"TestMessage","definitions":[{"name":"TestString","isOptional":"false","value":{"isScalar":"true","value":"string"}}]}]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input.ToJSON())
		})
	}
}