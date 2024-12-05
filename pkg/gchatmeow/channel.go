package gchatmeow

import (
	_ "bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	_ "unicode/utf16"
	"unicode/utf8"

	"go.mau.fi/util/pblite"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

const (
	channelURLBase    = "https://chat.google.com/u/0/webchannel/"
	pushTimeout       = 60 * time.Second
	maxReadBytes      = 1024 * 1024
	lenRegexPattern   = `([0-9]+)\n`
	protocolVersion   = 8
	initialRequestRID = 10000
)

var (
	ErrNetworkError     = errors.New("network error")
	ErrUnexpectedStatus = errors.New("unexpected status")
	ErrSIDInvalid       = errors.New("SID invalid")
	ErrSIDExpiring      = errors.New("SID expiring")
	ErrChannelLifetime  = errors.New("channel lifetime expired")

	lenRegex = regexp.MustCompile(lenRegexPattern)
)

// Channel represents a BrowserChannel client
type Channel struct {
	session          *Session
	maxRetries       int
	retryBackoffBase int
	isConnected      bool
	onConnectCalled  bool
	chunkParser      *ChunkParser
	sidParam         string
	csessionidParam  string
	aid              int
	ofs              int
	rid              int

	OnConnect      *Event
	OnReconnect    *Event
	OnDisconnect   *Event
	OnReceiveArray *Event
}

// ChunkParser parses data from the backward channel into chunks
type ChunkParser struct {
	buf []byte
}

// NewChunkParser creates a new ChunkParser instance
func NewChunkParser() *ChunkParser {
	return &ChunkParser{
		buf: make([]byte, 0),
	}
}

// bestEffortDecode attempts to decode as much UTF-8 data as possible from the buffer
func bestEffortDecode(data []byte) string {
	valid := make([]byte, 0, len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError {
			break
		}
		valid = append(valid, data[:size]...)
		data = data[size:]
	}
	return string(valid)
}

// GetChunks yields chunks generated from received data.
// The buffer may not be decodable as UTF-8 if there's a split multi-byte
// character at the end. To handle this, we do a "best effort" decode of the
// buffer to decode as much of it as possible.
func (p *ChunkParser) GetChunks(newDataBytes []byte) []string {
	var chunks []string
	p.buf = append(p.buf, newDataBytes...)

	for {
		// Decode buffer with best effort
		bufDecoded := bestEffortDecode(p.buf)

		// Convert to UTF-16 (removing BOM)
		var bufUtf16 []byte
		for _, r := range bufDecoded {
			// Convert each rune to UTF-16
			buf := make([]byte, 2)
			binary.BigEndian.PutUint16(buf, uint16(r))
			bufUtf16 = append(bufUtf16, buf...)
		}

		// Find length string match
		matches := lenRegex.FindStringSubmatch(bufDecoded)
		if matches == nil {
			break
		}

		lengthStr := matches[1]
		// Both lengths are in number of bytes in UTF-16 encoding
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			break
		}
		length *= 2 // Convert to UTF-16 byte count

		// Calculate length of the submission length and newline in UTF-16
		lenStrAndNewline := lengthStr + "\n"
		var lenLength int
		for _, r := range lenStrAndNewline {
			lenLength += 2 // Each UTF-16 character is 2 bytes
			_ = r
		}

		if len(bufUtf16)-lenLength < length {
			break
		}

		// Extract submission
		submission := bufUtf16[lenLength : lenLength+length]

		// Convert UTF-16 bytes back to string
		var result string
		for i := 0; i < len(submission); i += 2 {
			if i+1 >= len(submission) {
				break
			}
			char := binary.BigEndian.Uint16(submission[i : i+2])
			result += string(rune(char))
		}

		chunks = append(chunks, result)

		// Calculate how many bytes to drop from the buffer
		dropLength := len(matches[0]) // length of the length string and newline
		dropLength += len(result)     // length of the actual content in UTF-8

		if dropLength <= len(p.buf) {
			p.buf = p.buf[dropLength:]
		} else {
			p.buf = p.buf[:0]
		}
	}

	return chunks
}

func NewChannel(session *Session, maxRetries, retryBackoffBase int) (*Channel, error) {

	return &Channel{
		session:          session,
		maxRetries:       maxRetries,
		retryBackoffBase: retryBackoffBase,
		rid:              initialRequestRID + rand.Intn(89999),
		OnConnect:        &Event{},
		OnReconnect:      &Event{},
		OnDisconnect:     &Event{},
		OnReceiveArray:   &Event{},
	}, nil
}

// IsConnected returns whether the channel is currently connected
func (c *Channel) IsConnected() bool {
	return c.isConnected
}

// Listen listens for messages on the backwards channel
func (c *Channel) Listen(ctx context.Context, maxAge time.Duration) error {
	retries := 0
	skipBackoff := false

	csessionidParam, err := c.register(ctx)
	if err != nil {
		return fmt.Errorf("failed to register: %w", err)
	}
	c.csessionidParam = csessionidParam

	start := time.Now()

	for retries <= c.maxRetries {
		if time.Since(start) > maxAge {
			return ErrChannelLifetime
		}

		if retries > 0 && !skipBackoff {
			backoffSeconds := time.Duration(pow(c.retryBackoffBase, retries)) * time.Second
			log.Printf("Backing off for %v seconds", backoffSeconds)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoffSeconds):
			}
		}
		skipBackoff = false

		c.chunkParser = &ChunkParser{}

		err := c.longPollRequest(ctx)
		if err != nil {
			if errors.Is(err, ErrSIDExpiring) {
				log.Printf("Long-polling interrupted: %v", err)

				csessionidParam, err = c.register(ctx)
				if err != nil {
					return fmt.Errorf("failed to re-register: %w", err)
				}
				c.csessionidParam = csessionidParam

				retries++
				skipBackoff = true
				continue
			}

			log.Printf("Long-polling request failed: %v", err)
			retries++

			if c.isConnected {
				c.isConnected = false
				c.OnDisconnect.Fire(nil)
			}

			continue
		}

		retries = 0
	}

	return fmt.Errorf("ran out of retries for long-polling request")
}

// register registers a new channel and returns the csessionid
func (c *Channel) register(ctx context.Context) (string, error) {
	c.sidParam = ""
	c.aid = 0
	c.ofs = 0

	resp, err := c.session.FetchRaw(ctx, "GET", channelURLBase+"register?ignore_compass_cookie=1", nil, http.Header{
		"Content-Type": {"application/x-protobuf"},
	}, true, nil)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("%w: %d %s - %s", ErrUnexpectedStatus, resp.StatusCode, resp.Status, string(body))
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "COMPASS" {
			if strings.HasPrefix(cookie.Value, "dynamite-ui=") {
				return cookie.Value[len("dynamite-ui="):], nil
			}
			log.Printf("COMPASS cookie doesn't start with dynamite-ui= (value: %s)", cookie.Value)
		}
	}

	return "", nil
}

func (c *Channel) sendStreamEvent(ctx context.Context, request proto.StreamEventsRequest) error {
	params := url.Values{
		"VER": []string{"8"},                      // channel protocol version
		"RID": []string{fmt.Sprintf("%d", c.rid)}, // request identifier
		"t":   []string{"1"},                      // trial
		"SID": []string{c.sidParam},               // session ID
		"AID": []string{string(c.aid)},            // last acknowledged id
	}
	c.rid++

	// Prepare headers
	headers := http.Header{
		"Content-Type": []string{"application/x-www-form-urlencoded"},
	}

	protoBytes, err := pblite.Marshal(&request)
	if err != nil {
		return fmt.Errorf("failed to marshal events request: %w", err)
	}

	jsonBody, err := json.Marshal(protoBytes)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON body: %w", err)
	}

	// Prepare form data
	formData := url.Values{
		"count":     []string{"1"},
		"ofs":       []string{fmt.Sprintf("%d", c.ofs)},
		"req0_data": []string{string(jsonBody)},
	}
	c.ofs++

	res, err := c.session.FetchRaw(ctx, "POST", channelURLBase+"events", params, headers, true, []byte(formData.Encode()))
	if err != nil {
		return err
	}
	_ = res
	return nil
}

func (c *Channel) sendInitialPing(ctx context.Context) error {
	state := proto.PingEvent_ACTIVE
	focusState := proto.PingEvent_FOCUS_STATE_FOREGROUND
	interactiveState := proto.PingEvent_INTERACTIVE
	notificationsEnabled := true
	event := proto.PingEvent{State: &state, ApplicationFocusState: &focusState, ClientInteractiveState: &interactiveState, ClientNotificationsEnabled: &notificationsEnabled}
	return c.sendStreamEvent(ctx, proto.StreamEventsRequest{PingEvent: &event})
}

// longPollRequest opens a long-polling request and receives arrays
func (c *Channel) longPollRequest(ctx context.Context) error {
	params := url.Values{
		"VER": {strconv.Itoa(protocolVersion)},
		"RID": {strconv.Itoa(c.rid)},
		"t":   {"1"},
		"zx":  {uniqueID()},
	}

	if c.sidParam == "" {
		params.Set("CVER", "22")
		params.Set("$req", "count=1&ofs=0&req0_data=%5B%5D")
		params.Set("SID", "null")
		c.rid++
	} else {
		params.Set("CI", "0")
		params.Set("TYPE", "xmlhttp")
		params.Set("RID", "rpc")
		params.Set("AID", strconv.Itoa(c.aid))
		params.Set("SID", c.sidParam)
	}

	resp, err := c.session.FetchRaw(ctx, "GET", channelURLBase+"events", params, http.Header{
		"referer": {"https://chat.google.com/"},
	}, true, nil)
	if err != nil {
		if err == context.DeadlineExceeded {
			return fmt.Errorf("%w: request timed out", ErrNetworkError)
		}
		return fmt.Errorf("%w: %v", ErrNetworkError, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusBadRequest {
			if resp.Status == "Unknown SID" || strings.Contains(string(body), "Unknown SID") {
				return ErrSIDInvalid
			}
		}
		return fmt.Errorf("%w: %d %s - %s", ErrUnexpectedStatus, resp.StatusCode, resp.Status, string(body))
	}

	if initialResp := resp.Header.Get("X-HTTP-Initial-Response"); initialResp != "" {
		sid, err := parseSIDResponse(initialResp)
		if err != nil {
			return fmt.Errorf("failed to parse SID response: %w", err)
		}

		if c.sidParam != sid {
			c.sidParam = sid
			c.aid = 0
			c.ofs = 0

			params := url.Values{
				"VER":  []string{"8"},
				"RID":  []string{"rpc"},
				"SID":  []string{c.sidParam},
				"AID":  []string{string(c.aid)},
				"CI":   []string{"0"},
				"TYPE": []string{"xmlhttp"},
				"zx":   []string{uniqueID()},
				"t":    []string{"1"},
			}

			if _, err := c.session.FetchRaw(ctx, "GET", channelURLBase+"events", params, nil, true, nil); err != nil {
				return fmt.Errorf("failed to acknowledge sid")
			}

			if err := c.sendInitialPing(ctx); err != nil {
				return fmt.Errorf("failed to send initial ping: %w", err)
			}
		}
	}

	reader := resp.Body
	buffer := make([]byte, maxReadBytes)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			n, err := reader.Read(buffer)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				if strings.Contains(err.Error(), "use of closed network connection") {
					return ErrSIDExpiring
				}
				return fmt.Errorf("%w: %v", ErrNetworkError, err)
			}

			if err := c.onPushData(buffer[:n]); err != nil {
				return fmt.Errorf("failed to process push data: %w", err)
			}
		}
	}
}

// OnPushData handles incoming push data and triggers appropriate events
func (c *Channel) onPushData(dataBytes []byte) error {
	// Log received chunk
	log.Printf("Received chunk:\n%s", string(dataBytes))

	// Update metrics
	// c.receivedChunks.Add(float64(len(dataBytes)))

	// Process chunks
	chunks := c.chunkParser.GetChunks(dataBytes)

	for _, chunk := range chunks {
		// Handle connection state
		if !c.isConnected {
			if c.onConnectCalled {
				c.isConnected = true
				c.OnReconnect.Fire(nil)
			} else {
				c.onConnectCalled = true
				c.isConnected = true
				c.OnConnect.Fire(nil)
			}
		}

		// Parse the container array
		var containerArray [][]interface{}
		if err := json.Unmarshal([]byte(chunk), &containerArray); err != nil {
			fmt.Println("failed chunk:", chunk)
			return fmt.Errorf("failed to unmarshal chunk: %w", err)
		}

		// Process each inner array
		for _, innerArray := range containerArray {
			// Ensure the inner array has exactly 2 elements
			if len(innerArray) != 2 {
				return fmt.Errorf("invalid inner array length: expected 2, got %d", len(innerArray))
			}

			// Extract array ID and data array
			arrayID, ok := innerArray[0].(float64)
			if !ok {
				return fmt.Errorf("array ID is not a number")
			}

			dataArray := innerArray[1]

			log.Printf("Chunk contains data array with id %d:\n%v", arrayID, dataArray)

			// Fire receive array event
			c.OnReceiveArray.Fire(dataArray)

			// Update last array ID
			c.aid = int(math.Round(arrayID))
		}
	}

	return nil
}

// Helper functions (implementations not shown for brevity)
func uniqueID() string {
	// Implementation of _unique_id from Python code
	return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%d", rand.Int63())))
}

func parseSIDResponse(response string) (string, error) {
	var data []interface{}
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return "", err
	}

	if len(data) < 1 {
		return "", errors.New("invalid SID response format")
	}

	arr, ok := data[0].([]interface{})
	if !ok || len(arr) < 2 {
		return "", errors.New("invalid SID response array format")
	}

	sid, ok := arr[1].([]interface{})[1].(string)
	if !ok {
		return "", errors.New("invalid SID format in response")
	}

	return sid, nil
}

func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}
