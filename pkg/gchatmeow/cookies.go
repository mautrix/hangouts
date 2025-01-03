package gchatmeow

import (
	"reflect"
)

type Cookies struct {
	COMPASS string
	SSID    string
	SID     string
	OSID    string
	HSID    string
}

var (
	CookieNames = []string{"COMPASS", "SSID", "SID", "OSID", "HSID"}
)

func (c *Cookies) UpdateValues(values map[string]string) {
	r := reflect.ValueOf(c)
	for _, key := range CookieNames {
		field := reflect.Indirect(r).FieldByName(key)
		field.SetString(values[key])
	}
}
