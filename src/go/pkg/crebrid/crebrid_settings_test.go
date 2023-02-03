package crebrid

import (
	"reflect"
	"testing"
)

func TestLoadFromByteArr(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		want    *CrebridDSettings
		wantErr bool
	}{
		{
			name: "valid complete config data",
			args: args{
				data: "# this a comment\nip=192.123.45.67\nport=65432\nipcPort=76543\naccessCode=123DEF",
			},
			want: &CrebridDSettings{
				IP:         "192.123.45.67",
				Port:       65432,
				IPCPort:    76543,
				AccessCode: "123DEF",
			},
			wantErr: false,
		},
		{
			name: "valid config data w/o IPC port",
			args: args{
				data: "# this a comment\nip=192.123.45.67\nport=41296",
			},
			want: &CrebridDSettings{
				IP:         "192.123.45.67",
				Port:       41296,
				IPCPort:    65432,
				AccessCode: "123ABC",
			},
			wantErr: false,
		},
		{
			name: "valid config data w/o IPC port",
			args: args{
				data: "# this a comment\nip 192.123.45.67\nport=41296",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadFromByteArr([]byte(tt.args.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromByteArr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadFromByteArr() = %v, want %v", got, tt.want)
			}
		})
	}
}
