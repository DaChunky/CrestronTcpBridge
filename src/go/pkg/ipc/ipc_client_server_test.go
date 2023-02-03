package ipc

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func ipcEventHandler(cc *ClientCommand) (*ServerResponse, error) {
	sr := NewServerResponse()
	sr.Cmd = cc.Cmd
	switch cc.Cmd {
	case IC_REGISTER:
		sr.ID = uuid.NewString()
	case IC_SINGLE:
		sr.ID = cc.ID
		sr.DigitalPortInfo[cc.DigitalPorts[0]] = true
	case IC_MULTIPLE:
		sr.ID = cc.ID
		for _, port := range cc.DigitalPorts {
			sr.DigitalPortInfo[port] = true
		}
	}
	return sr, nil
}

func TestIpcClientServer(t *testing.T) {
	tests := []struct {
		name    string
		cc      ClientCommand
		wantErr bool
	}{
		{
			name: "single command",
			cc: ClientCommand{
				Cmd:          IC_SINGLE,
				ID:           "client123",
				DigitalPorts: []int{3},
			},
			wantErr: false,
		},
		{
			name: "multiple command",
			cc: ClientCommand{
				Cmd:          IC_MULTIPLE,
				ID:           "client123",
				DigitalPorts: []int{2, 3, 9},
			},
			wantErr: false,
		},
	}
	port := 65432
	srv := NewIpcServer(port)
	go srv.StartListening(ipcEventHandler)
	defer srv.Close()
	time.Sleep(100 * time.Millisecond)
	err := srv.HasError()
	if err != nil {
		t.Fatalf("failed to listening to port [%d]: %v", port, err)
	}
	client, err := RegisterClient("localhost", port)
	if err != nil {
		t.Fatalf("failed to register client on port [%d]: %v", port, err)
	}
	defer client.CloseConnection()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &tt.cc
			sr, err := client.SendCommand(cc)
			if (err != nil) != tt.wantErr {
				t.Fatalf("failed to send command on port [%d]: %v", port, err)
			}
			if sr.Cmd != cc.Cmd {
				t.Fatalf("sr.Cmd = %d: want %d", sr.Cmd, cc.Cmd)
			}
			for _, port := range cc.DigitalPorts {
				if _, ok := sr.DigitalPortInfo[port]; !ok {
					t.Fatalf("want port %d but don't receive is", port)
				}
			}
		})
	}
	t.Log("--> tests run")
}
