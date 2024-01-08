package security

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	type args struct {
		password string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "should successfully generate hashed password",
			args:    args{password: "XXXXXXXX"},
			wantErr: false,
		},
		{
			name:    "should fail to generate hashed password because of empty string password",
			args:    args{password: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HashPassword(tt.args.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got == "" {
				t.Errorf("HashPassword(), expected non-empty string, got empty string")
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	hashedPassword, _ := HashPassword("password123")
	type args struct {
		hashPassword string
		password     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "should successfully validate password",
			args: args{
				hashPassword: hashedPassword,
				password:     "password123",
			},
			want: true,
		},
		{
			name: "should fail to validate password",
			args: args{
				hashPassword: hashedPassword,
				password:     "password1234",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidatePassword(tt.args.hashPassword, tt.args.password); got != tt.want {
				t.Errorf("ValidatePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
