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
		input    []byte
		expected *Device
		WantErr  bool
	}{
		{
			Name: "Default test",
			input: []byte(`
				message TestMessage{string TestString;};service TestService(TestMessage):string;
			`),
			expected: &Device{
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 80,
				},
				Services: []serv.Service{
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
				Messages: []serv.Message{
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
			Name: "One Message",
			input: []byte(`
				message TestMessage{string TestString;};
			`),
			expected: &Device{
				Addr: &net.TCPAddr{
					IP:   net.IPv4(127, 0, 0, 1),
					Port: 80,
				},
				Messages: []serv.Message{
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
			dev, err := NewDevice(addr, tt.input)
			if tt.WantErr {
				assert.Error(t, err)
			}
			tt.expected.ID = dev.ID
			assert.Equal(t, tt.expected, dev)
		})
	}
}
