package crebrid

import (
	"reflect"
	"testing"
)

func TestSystemStatusFromJSON(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    *SystemStatus
		wantErr bool
	}{
		{
			name: "valid system status",
			args: args{
				str: "{\"d\":[1,0,1,0,0],\"a\":[23.4,86,21.7]}",
			},
			want: &SystemStatus{
				D: []int{1, 0, 1, 0, 0},
				A: []float64{23.4, 86, 21.7},
			},
			wantErr: false,
		},
		{
			name: "invalid system status",
			args: args{
				str: "{\"d\":[1,0,1,0,0,\"a\":[23.4,86,21.7]}",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SystemStatusFromJSON(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("SystemStatusFromJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SystemStatusFromJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
