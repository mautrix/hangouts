package gchatmeow

import (
	"reflect"
)

type Cookies struct {
	Compass string
	SSID    string
	SID     string
	OSID    string
	HSID    string
}

var (
	cookies = []string{"Compass", "SSID", "SID", "OSID", "HSID"}
)

func (c *Cookies) UpdateValues(values map[string]string) {
	r := reflect.ValueOf(c)
	for _, key := range cookies {
		field := reflect.Indirect(r).FieldByName(key)
		field.SetString(values[key])
	}
}
