package models

type Settings struct {
	AuthTokenSecret   string
	AuthTokenDuration int64
}

func NewSettings() *Settings {
	return &Settings{
		AuthTokenSecret:   "test",
		AuthTokenDuration: 900,
	}
}
