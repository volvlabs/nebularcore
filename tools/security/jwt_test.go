package security

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
)

func TestNewJWT(t *testing.T) {
	type args struct {
		payload         jwt.MapClaims
		signingKey      string
		secondsDuration int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"should successfully generate token",
			args{jwt.MapClaims{}, "test", 31536000},
			false,
		},
		{
			"should return successfully even with payload",
			args{jwt.MapClaims{"test": "test"}, "test", 10},
			false,
		},
		{
			"should return an error because of zero duration passed",
			args{jwt.MapClaims{}, "test", 0},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewJWT(tt.args.payload, tt.args.signingKey, tt.args.secondsDuration)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && got == "" {
				t.Errorf("NewJWT() returned is empty")
			}
		})
	}
}

func TestParseJWT(t *testing.T) {
	type args struct {
		token           string
		verificationKey string
	}
	tests := []struct {
		name    string
		args    args
		want    jwt.MapClaims
		wantErr bool
	}{
		{
			"should successfully parse the token",
			args{
				token:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoidGVzdCIsImV4cCI6MTg5ODYzNjEzN30.gqRkHjpK5s1PxxBn9qPaWEWxTbpc1PPSD-an83TsXRY",
				verificationKey: "test",
			},
			jwt.MapClaims{"name": "test", "exp": 1898636137.0},
			false,
		},
		{
			"should return error as token is already expired",
			args{
				token:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MDM4NDk4MTgsInRlc3QiOiJ0ZXN0In0.gIgTx5u-kOB7YAxp9ppEpKPXuSeDx1n_0wEq4FBQ8wU",
				verificationKey: "test",
			},
			nil,
			true,
		},
		{
			"should return an error because of invalid token",
			args{
				token:           "invalid_token",
				verificationKey: "test",
			},
			nil,
			true,
		},
		{
			"should return an error because of invalid verification key",
			args{
				token:           "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MzUzODUxMTUsInRlc3QiOiJ0ZXN0In0.o2fePB78FKl4BQWEugCYGuvWeguAOWwXn17EnN3hQcM",
				verificationKey: "invalid_key",
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJWT(tt.args.token, tt.args.verificationKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJWT() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for k, v := range tt.want {
				v2, ok := got[k]
				if !ok {
					t.Errorf("ParseJWT() key %s not found in result", k)
				}
				if v != v2 {
					t.Errorf("ParseJWT() want %v for %s claim, got %v", v, k, v2)
				}
			}
		})
	}
}
