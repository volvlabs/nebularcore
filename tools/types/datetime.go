package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cast"
)

const DefaultDateFormat = "2006-01-02 15:04:05"

func NowDateTime() DateTime {
	return DateTime{t: time.Now().UTC()}
}

type DateTime struct {
	t time.Time
}

func ParseDateTime(value any) (DateTime, error) {
	d := DateTime{}
	err := d.Scan(value)
	return d, err
}

func (d DateTime) Time() time.Time {
	return d.t
}

func (d DateTime) Value() (driver.Value, error) {
	return d.t, nil
}

func (d *DateTime) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		d.t = v
		return nil
	case DateTime:
		d.t = v.t
	case string:
		if v == "" {
			d.t = time.Time{}
		} else {
			t, err := time.Parse(DefaultDateFormat, v)
			if err != nil {
				t = cast.ToTime(v)
			}
			d.t = t
		}
	case int:
		t := time.UnixMilli(int64(v))
		d.t = t
	case int64:
		t := time.UnixMilli(v)
		d.t = t
	case float32:
		t := time.UnixMilli(int64(v))
		d.t = t
	case float64:
		t := time.UnixMilli(int64(v))
		d.t = t
	default:
		str := cast.ToString(v)
		if str == "" {
			d.t = time.Time{}
		} else {
			d.t = cast.ToTime(str)
		}
	}
	return nil
}

func (d DateTime) String() string {
	return d.Time().UTC().Format(DefaultDateFormat)
}

func (d DateTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", d.Time().UnixMilli())), nil
}

func (d *DateTime) UnmarshalJSON(b []byte) error {
	var s string
	var err error
	if err = json.Unmarshal(b, &s); err == nil {
		return d.Scan(s)
	}

	var n int64
	if err = json.Unmarshal(b, &n); err == nil {
		return d.Scan(n)
	}

	return err
}
