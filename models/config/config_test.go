package config

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		configFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    *AppConfig
		wantErr bool
	}{
		{
			name: "should return the app config",
			args: args{configFilePath: "../../test/data/config.yml"},
			want: &AppConfig{
				IsDev:      true,
				EnforceAcl: true,
				Server: ServeConfig{
					Port:            "8080",
					Host:            "0.0.0.0",
					ShutdownTimeout: 5,
					AllowedOrigins:  "*",
				},
				Database: DatabaseConfig{
					Host:     "postgres",
					Username: "postgres",
					Password: "password",
					Name:     "postgres",
					Port:     "5432",
					SSLMode:  "disable",
				},
			},
			wantErr: false,
		},
		{
			name:    "should return error since file does not exist",
			args:    args{configFilePath: "../../test/data/invalid_config.yaml"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "should return empty app config when config file is not passed",
			args:    args{configFilePath: ""},
			want:    &AppConfig{},
			wantErr: false,
		},
		{
			name:    "should return error since config file is empty",
			args:    args{configFilePath: "../../test/data/empty_config.yaml"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.configFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
