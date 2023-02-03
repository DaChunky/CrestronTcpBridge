package ipc

import (
	"reflect"
	"testing"
)

func TestClientCommand(t *testing.T) {
	type fields struct {
		Cmd          int
		ID           string
		DigitalPorts []int
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "register client command",
			fields: fields{
				Cmd:          IC_REGISTER,
				ID:           "",
				DigitalPorts: []int{},
			},
			wantErr: false,
		},
		{
			name: "single client command",
			fields: fields{
				Cmd:          IC_SINGLE,
				ID:           "client123",
				DigitalPorts: []int{2},
			},
			wantErr: false,
		},
		{
			name: "multiple client command",
			fields: fields{
				Cmd:          IC_MULTIPLE,
				ID:           "client123",
				DigitalPorts: []int{2, 6, 7},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &ClientCommand{
				Cmd:          tt.fields.Cmd,
				ID:           tt.fields.ID,
				DigitalPorts: tt.fields.DigitalPorts,
			}
			got, err := cc.GetCommand2Send()
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientCommand.GetCommand2Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			cc2, err := ClientCommandFromRequest(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientCommandFromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(cc2, cc) {
				t.Errorf("ClientCommand = %v, want %v", cc2, cc)
			}
		})
	}
}

func TestClientCommand_AddDigitalPorts(t *testing.T) {
	type fields struct {
		Cmd          int
		ID           string
		DigitalPorts []int
	}
	type args struct {
		ports []int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		expect []int
	}{
		{
			name: "add ports",
			fields: fields{
				Cmd:          IC_MULTIPLE,
				ID:           "client123",
				DigitalPorts: []int{2, 6},
			},
			args: args{
				ports: []int{7, 9},
			},
			expect: []int{2, 6, 7, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &ClientCommand{
				Cmd:          tt.fields.Cmd,
				ID:           tt.fields.ID,
				DigitalPorts: tt.fields.DigitalPorts,
			}
			cc.AddDigitalPorts(tt.args.ports...)
			if !reflect.DeepEqual(tt.expect, cc.DigitalPorts) {
				t.Fatalf("DigitalPorts = %v, want %v", cc.DigitalPorts, tt.expect)
			}
		})
	}
}

func TestServerResponse(t *testing.T) {
	type fields struct {
		Cmd             int
		ID              string
		DigitalPortInfo map[int]bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "register response",
			fields: fields{
				Cmd:             IC_REGISTER,
				ID:              "client123",
				DigitalPortInfo: map[int]bool{},
			},
			wantErr: false,
		},
		{
			name: "single response",
			fields: fields{
				Cmd: IC_SINGLE,
				ID:  "client123",
				DigitalPortInfo: map[int]bool{
					2: true,
				},
			},
			wantErr: false,
		},
		{
			name: "single response",
			fields: fields{
				Cmd: IC_MULTIPLE,
				ID:  "client123",
				DigitalPortInfo: map[int]bool{
					2: true,
					3: false,
					6: true,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := &ServerResponse{
				Cmd:             tt.fields.Cmd,
				ID:              tt.fields.ID,
				DigitalPortInfo: tt.fields.DigitalPortInfo,
			}
			got, err := sr.GetResponse2Send()
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerResponse.GetResponse2Send() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sr2, err := ServerResponseFromResponse(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ServerResponseFromResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(sr, sr2) {
				t.Errorf("ServerResponse.GetResponse2Send() = %v, want %v", sr2, sr)
			}
		})
	}
}
