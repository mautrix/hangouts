package query

import (
	"strings"

	urlquery "github.com/google/go-querystring/query"
	"go.mau.fi/util/random"
)

type WebChannelEventsQuery struct {
	Ver  string `url:"VER,omitempty"`
	CVer string `url:"CVER,omitempty"`
	Rid  string `url:"RID,omitempty"`
	Sid  string `url:"SID,omitempty"`
	Req  string `url:"$req,omitempty"`
	Aid  string `url:"AID,omitempty"`
	Ci   string `url:"CI,omitempty"`
	Type string `url:"TYPE,omitempty"`
	Zx   string `url:"zx,omitempty"`
	T    string `url:"t,omitempty"`
	Rt   string `url:"rt,omitempty"`
}

func (q *WebChannelEventsQuery) Encode() (string, error) {
	q.Zx = strings.ToLower(random.String(12))
	v, err := urlquery.Values(q)
	if err != nil {
		return "", err
	}
	return v.Encode(), nil
}

type WebChannelPingQuery struct {
	Count   int    `url:"count"`
	Ofs     int    `url:"ofs"`
	ReqData string `url:"req0_data"`
}

func (q *WebChannelPingQuery) Encode() (string, error) {
	v, err := urlquery.Values(q)
	if err != nil {
		return "", err
	}
	return v.Encode(), nil
}
