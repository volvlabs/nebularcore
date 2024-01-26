package models

type Settings struct {
	AuthTokenSecret     string
	OtpGenerationSecret string
	OtpPeriod           uint
	AuthTokenDuration   int64
}

func NewSettings() *Settings {
	return &Settings{
		AuthTokenSecret:     "test",
		OtpGenerationSecret: "otp_secret",
		OtpPeriod:           900,
		AuthTokenDuration:   900,
	}
}
