package gchatmeow

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/endpoints"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/query"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/event"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/pblite"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatproto"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatprotoweb"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/types"
)

var LEN_REGEX = regexp.MustCompile(`[0-9]+\n`)
var MAX_READ_BYTES = 1024 * 1024

type RealtimeClient struct {
	client     *Client
	http       *http.Client
	conn       *http.Response
	cancelFunc context.CancelFunc

	pingTimer    *time.Ticker
	pingInterval int
	session      *query.WebChannelEventsQuery
	reqId        int
	pingReqId    int
	ackId        int
}

func (c *Client) newRealtimeClient() *RealtimeClient {
	return &RealtimeClient{
		client: c,
		http: &http.Client{
			Transport: &http.Transport{
				Proxy: c.httpProxy,
			},
		},
		session: &query.WebChannelEventsQuery{
			Ver: "8",
			Sid: "null",
			T:   "1",
		},
		ackId: 1,
		reqId: 0,
	}
}

func (rc *RealtimeClient) Connect() error {
	if rc.session == nil || rc.session.Sid == "null" {
		err := rc.Register()
		if err != nil {
			return err
		}
	}

	queryArgs, err := rc.getEventsQueryString(false)
	if err != nil {
		return err
	}

	url := endpoints.WEBCHANNEL_EVENTS + "?" + queryArgs

	ctx, cancel := context.WithCancel(context.Background())
	rc.cancelFunc = cancel

	headers := rc.client.buildHeaders(types.HeaderOpts{
		WithCookies:     true,
		WithXClientData: true,
	})
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header = headers

	conn, err := rc.http.Do(req)
	if err != nil {
		return err
	}

	if conn.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", conn.Status)
	}

	connHeaders := conn.Header
	connSessionDataStr := connHeaders.Get("X-Http-Initial-Response")
	if connSessionDataStr == "" && rc.session.Sid == "null" {
		return fmt.Errorf("failed to retrieve session data while trying to connect to realtime server")
	}

	if rc.session.Sid == "null" {
		connSessionData := &gchatprotoweb.WebChannelInitialHttpSessionData{}
		err = pblite.Unmarshal([]byte(connSessionDataStr), connSessionData)
		if err != nil {
			return err
		}

		rc.pingInterval = int(connSessionData.Data.Session.PingInterval)
		rc.session.Sid = connSessionData.Data.Session.SessionId
	}

	rc.conn = conn
	go rc.beginReadStream()
	rc.startPinging()

	return nil
}

func (rc *RealtimeClient) Reconnect(reason string) {
	rc.client.Logger.Info().Str("reason", reason).Msg("Attempting to reconnect to webchannel connection...")
	rc.pingTimer.Stop()
	err := rc.Connect()
	if err != nil {
		log.Fatalf("failed to reconnect after disconnecting due to %s error (%s)", reason, err.Error())
	}
}

func (rc *RealtimeClient) updateAckID(arrayId int) {
	if arrayId > rc.ackId {
		rc.ackId = arrayId
	}
}

func (rc *RealtimeClient) startPinging() {
	if rc.pingTimer != nil {
		rc.pingTimer.Stop()
	}
	rc.pingTimer = time.NewTicker(time.Duration(rc.pingInterval) * time.Millisecond)
	go rc.beginPinging()
}

func (rc *RealtimeClient) beginPinging() {
	err := rc.sendPing()
	if err != nil {
		rc.client.Logger.Err(err).Msg("Failed to send ping request to webchannel")
	}
	for range rc.pingTimer.C {
		err = rc.sendPing()
		if err != nil {
			rc.client.Logger.Err(err).Msg("Failed to send ping request to webchannel")
		}
	}
}

func (rc *RealtimeClient) beginReadStream() {
	reader := bufio.NewReader(rc.conn.Body)
	for {
		data := make([]byte, MAX_READ_BYTES)
		n, err := reader.Read(data)
		if err != nil {
			if err == io.EOF {
				defer func() {
					go rc.Reconnect("EOF")
				}()
				break
			}
			log.Fatalf("failed to read bytes: %s", err.Error())
		}

		fullData := data[:n]
		splitData := LEN_REGEX.Split(string(fullData), -1)

		pbliteDataStr := splitData[1]

		var arrayData []any
		err = json.Unmarshal([]byte(pbliteDataStr), &arrayData)
		if err != nil {
			log.Fatalf("failed to unmarshal pblite data received from realtime stream into []any (%s) (%s)", pbliteDataStr, err.Error())
		}

		for _, array := range arrayData {
			arraySlice := array.([]any)
			arr := arraySlice[1].([]any)
			switch arr[0].(type) {
			case []any:
				eventData := &gchatproto.WebChannelEventMetadata{}
				err = pblite.UnmarshalSlice(arraySlice, eventData)
				if err != nil {
					log.Fatalf("failed to parse pblite data: %s", err.Error())
				}
				rc.processEvent(eventData)
			case string:
				// "noop" events are received here, idk what they are for though
				break
			}
		}
	}
}

func (rc *RealtimeClient) getEventsQueryString(isPing bool) (string, error) {
	eventsQuery := query.WebChannelEventsQuery{
		Ver: "8",
		Sid: rc.session.Sid,
		T:   "1",
	}

	if rc.session.Sid == "null" {
		// hardcoded data = fine ?
		// req0_data is RequestHeader
		eventsQuery.Req = "count=1&ofs=0&req0_data=%5Bnull%2Cnull%2Cnull%2Cnull%2C%5B8%2C159%2C%22web-q8%2FvzXPOdUeOoRD6rNwBYo%2FsHy4%3D%22%2C%5Bnull%2Cnull%2Cnull%2Cnull%2C2%2C2%2C2%2C2%2C2%2C2%2Cnull%2Cnull%2Cnull%2Cnull%2Cnull%2C2%2C2%2C2%2C2%2C2%2Cnull%2C2%2C2%2C2%2C2%2C2%2C2%2Cnull%2C2%2C0%2C0%2C2%2C2%2Cnull%2Cnull%2C0%2C0%2C0%2C0%2C0%2C0%2C2%2C0%2C0%5D%2C%22en%22%5D%2C5285528534243897%5D"
		eventsQuery.Rid = strconv.Itoa(rc.reqId)
		eventsQuery.CVer = "22"
	} else if !isPing {
		eventsQuery.Rid = "rpc"
		eventsQuery.Ci = "0"
		eventsQuery.Type = "xmlhttp"
		eventsQuery.Aid = strconv.Itoa(rc.ackId)
	}

	if isPing {
		eventsQuery.Ci = ""
		eventsQuery.CVer = ""
		eventsQuery.Aid = strconv.Itoa(rc.ackId)
	}

	return eventsQuery.Encode()
}

func (rc *RealtimeClient) sendPing() error {
	queryArgs, err := rc.getEventsQueryString(true)
	if err != nil {
		return err
	}

	pingPayload := query.WebChannelPingQuery{
		Count:   1,
		Ofs:     rc.pingReqId,
		ReqData: "[null,[2,null,null,null,3,0,null,null,0]]",
	}

	headers := rc.client.buildHeaders(types.HeaderOpts{
		WithCookies:     true,
		Referer:         "https://chat.google.com/",
		Origin:          "https://chat.google.com",
		WithXClientData: true,
	})

	encodedPayload, err := pingPayload.Encode()
	if err != nil {
		return err
	}

	url := endpoints.WEBCHANNEL_EVENTS + "?" + queryArgs
	req, respBody, err := rc.client.MakeRequest(url, "POST", headers, []byte(encodedPayload), types.FORM)
	if err != nil {
		return err
	}

	if req.StatusCode > 204 {
		return fmt.Errorf("could not ping webchannel connection (statusCode=%d, respBody=%s)", req.StatusCode, string(respBody))
	}

	rc.pingReqId += 1
	return nil
}

func (rc *RealtimeClient) Register() error {
	// rt = response_type = binary
	url := endpoints.WEBCHANNEL_REGISTER + "?ignore_compass_cookie=1"
	headers := rc.client.buildHeaders(types.HeaderOpts{
		WithCookies:     true,
		WithXClientData: true,
		Extra: map[string]string{
			"Referer": "https://chat.google.com/",
		},
	})

	resp, respBody, err := rc.client.MakeRequest(url, "GET", headers, nil, types.NONE)
	if err != nil {
		return err
	}

	if resp.StatusCode > 204 {
		return fmt.Errorf("failed to register webchannel session (statusCode=%d, respBody=%s)", resp.StatusCode, string(respBody))
	}

	rc.client.cookies.UpdateFromResponse(resp)

	return nil
}

func (rc *RealtimeClient) processEvent(channelEvent *gchatproto.WebChannelEventMetadata) {
	rc.updateAckID(int(channelEvent.ArrayId))
	eventData := channelEvent.DataWrapper.Data.Event
	for _, eventBody := range eventData.Bodies {
		prettifiedEventData := event.PrettifyEvent(eventBody)
		if prettifiedEventData == nil {
			rc.client.Logger.Fatal().Any("event_body", eventBody).Msg("Could not prettify event, unknown?")
		}

		if rc.client.eventHandler != nil {
			rc.client.eventHandler(prettifiedEventData)
		}
	}
}
