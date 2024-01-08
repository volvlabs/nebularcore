package types

import (
	"database/sql/driver"
	"time"
)

const DefaultDateFormat = "2006-01-02 15:04:05"

func NowDateTime() DateTime {
	return DateTime{t: time.Now().UTC()}
}

type DateTime struct {
	t time.Time
}

func (d DateTime) Time() time.Time {
	return d.t
}

func (d DateTime) Value() (driver.Value, error) {
	return d.t, nil
}

func (d *DateTime) Scan(value interface{}) error {
	if value == nil {
		d.t = time.Time{}
		return nil
	}

	d.t = value.(time.Time)
	return nil
}

func (d DateTime) String() string {
	return d.Time().UTC().Format(DefaultDateFormat)
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}
