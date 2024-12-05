package gchatmeow

import (
	_ "bytes"
	"context"
	_ "encoding/base64"
	"encoding/json"
	"fmt"
	_ "io"
	"log"
	_ "math/rand"
	_ "mime"
	"net/http"
	"net/url"
	"os"
	"regexp"
	_ "strings"
	"sync"
	"time"

	"go.mau.fi/util/pblite"
	"go.mau.fi/util/ptr"
	pb "google.golang.org/protobuf/proto"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

const (
	uploadURL = "https://chat.google.com/uploads"
	apiKey    = "AIzaSyD7InnYR3VKdb4j2rMUEbTCIr2VyEazl6k"
	gcBaseURL = "https://chat.google.com/u/0"
)

var (
	wizPattern = regexp.MustCompile(`>window.WIZ_global_data = ({.+?});</script>`)
	logger     = log.New(os.Stdout, "client: ", log.LstdFlags)
)

// Client represents an instant messaging client for Google Chat
type Client struct {
	session          *Session
	channel          *Channel
	maxRetries       int
	retryBackoffBase int

	// Events
	OnConnect     *Event
	OnReconnect   *Event
	OnDisconnect  *Event
	OnStreamEvent *Event

	// State
	gcRequestHeader   *proto.RequestHeader
	clientID          string
	email             string
	lastActiveSecs    float64
	activeClientState int
	apiReqID          int64
	xsrfToken         string
	lastTokenRefresh  float64

	// Mutex for thread safety
	mu sync.RWMutex
}

// NewClient creates a new Google Chat client
func NewClient(cookies *Cookies, userAgent string, maxRetries, retryBackoffBase int) *Client {
	if maxRetries == 0 {
		maxRetries = 5
	}
	if retryBackoffBase == 0 {
		retryBackoffBase = 2
	}

	session, err := NewSession(cookies, userAgent, os.Getenv("HTTP_PROXY"))

	if err != nil {
		return nil
	}
	c := &Client{
		maxRetries:       maxRetries,
		retryBackoffBase: retryBackoffBase,
		lastTokenRefresh: -86400,

		// Initialize channels for events
		OnConnect:     &Event{},
		OnReconnect:   &Event{},
		OnDisconnect:  &Event{},
		OnStreamEvent: &Event{},

		session: session,

		gcRequestHeader: &proto.RequestHeader{
			ClientType:    GetPointer(proto.RequestHeader_WEB),
			ClientVersion: GetPointer(int64(2440378181258)),
			ClientFeatureCapabilities: &proto.ClientFeatureCapabilities{
				SpamRoomInvitesLevel: GetPointer(proto.ClientFeatureCapabilities_FULLY_SUPPORTED),
			},
		},
	}

	return c
}

// Connect establishes a connection to the chat server
func (c *Client) Connect(ctx context.Context, maxAge time.Duration) error {
	c.apiReqID = 0

	if time.Now().Unix()-int64(c.lastTokenRefresh) > 86400 {
		logger.Println("Refreshing xsrf token before connecting")
		if err := c.RefreshTokens(ctx); err != nil {
			return fmt.Errorf("failed to refresh tokens: %w", err)
		}
	}

	channel, err := NewChannel(c.session, c.maxRetries, c.retryBackoffBase)
	if err != nil {
		return err
	}
	c.channel = channel

	// Set up event forwarding
	c.channel.OnConnect.AddObserver(func(interface{}) {
		c.OnConnect.Fire(nil)
	})
	c.channel.OnReceiveArray.AddObserver(c.onReceiveArray)
	// go func() {
	// 	for {
	// 		select {
	// 		case <-c.channel.OnConnect:
	// 			c.OnConnect <- struct{}{}
	// 		case <-c.channel.OnReconnect:
	// 			c.OnReconnect <- struct{}{}
	// 		case <-c.channel.OnDisconnect:
	// 			c.OnDisconnect <- struct{}{}
	// 		case arr := <-c.channel.OnReceiveArray:
	// 			if err := c.handleReceiveArray(arr); err != nil {
	// 				logger.Printf("Error handling receive array: %v", err)
	// 			}
	// 		case <-ctx.Done():
	// 			return
	// 		}
	// 	}
	// }()

	return c.channel.Listen(ctx, maxAge)
}

// refreshTokens makes a request to /mole/world to get required tokens
func (c *Client) RefreshTokens(ctx context.Context) error {
	params := url.Values{
		"origin": {"https://mail.google.com"},
		"shell":  {"9"},
		"hl":     {"en"},
		"wfi":    {"gtn-roster-iframe-id"},
		"hs":     {`["h_hs",null,null,[1,0],null,null,"gmail.pinto-server_20230730.06_p0",1,null,[15,38,36,35,26,30,41,18,24,11,21,14,6],null,null,"3Mu86PSulM4.en..es5",0,null,null,[0]]`},
	}

	headers := http.Header{
		"authority": {"chat.google.com"},
		"referer":   {"https://mail.google.com/"},
	}

	resp, err := c.session.Fetch(ctx, "GET", fmt.Sprintf("%s/mole/world", gcBaseURL), params, headers, true, nil)
	if err != nil {
		return err
	}

	os.WriteFile("body.html", resp.Body, 0644)
	// fmt.Println(string(resp.Body))

	matches := wizPattern.FindSubmatch(resp.Body)
	if len(matches) != 2 {
		return fmt.Errorf("didn't find WIZ_global_data in /mole/world response")
	}

	var wizData struct {
		QwAQke string `json:"qwAQke"`
		SMqcke string `json:"SMqcke"`
	}
	if err := json.Unmarshal(matches[1], &wizData); err != nil {
		return fmt.Errorf("non-JSON WIZ_global_data in /mole/world response: %w", err)
	}

	if wizData.QwAQke == "AccountsSignInUi" {
		// return ErrNotLoggedIn
		return fmt.Errorf("ErrNotLoggedIn")
	}

	fmt.Println("Done")
	c.mu.Lock()
	c.xsrfToken = wizData.SMqcke
	c.lastTokenRefresh = float64(time.Now().Unix())
	c.mu.Unlock()

	return nil
}

// OnReceiveArray parses channel array and calls appropriate events
func (c *Client) onReceiveArray(arg interface{}) {
	array, ok := arg.([]interface{})
	if !ok {
		fmt.Printf("expected arg to be []interface{}, got %T", array[0])
		return
	}

	if len(array) == 0 {
		fmt.Printf("received empty array")
		return
	}

	// Check for noop (keep-alive)
	if str, ok := array[0].(string); ok && str == "noop" {
		return // Ignore keep-alive
	}

	// Get the data from array
	data, ok := array[0].([]interface{})
	if !ok {
		fmt.Printf("expected array[0] to be []interface{}, got %T", array[0])
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create and decode protobuf response
	resp := &proto.StreamEventsResponse{}
	if err := pblite.Unmarshal(bytes, resp); err != nil {
		fmt.Printf("failed to decode proto: %w", err)
		return
	}

	fmt.Println(resp)

	// Process each event body
	for _, evt := range c.splitEventBodies(resp.GetEvent()) {
		log.Printf("Dispatching stream event: %v", evt)
		c.OnStreamEvent.Fire(evt)
	}

}

// SplitEventBodies splits an event with multiple bodies into separate events
func (c *Client) splitEventBodies(evt *proto.Event) []*proto.Event {
	if evt == nil {
		return nil
	}

	var events []*proto.Event

	// Handle embedded bodies
	embeddedBodies := evt.GetBodies()
	if len(embeddedBodies) > 0 {
		// Clear the bodies field in the original event
		evt.Bodies = nil
	}

	// If there's a body in the main event, include it
	if evt.Body != nil {
		events = append(events, pb.Clone(evt).(*proto.Event))
	}

	// Process each embedded body
	for _, body := range embeddedBodies {
		evtCopy := pb.Clone(evt).(*proto.Event)
		evtCopy.Body = pb.Clone(body).(*proto.Event_EventBody)
		evtCopy.Type = ptr.Ptr(body.GetEventType())
		events = append(events, evtCopy)
	}

	return events
}

func (c *Client) GetSelf(ctx context.Context) (*proto.User, error) {
	status, err := c.getSelfUserStatus(ctx)
	if err != nil {
		return nil, err
	}
	gcid := status.UserStatus.UserId.Id
	members, err := c.getMembers(ctx, gcid)
	if err != nil {
		return nil, err
	}
	return members.Members[0].GetUser(), nil
}

func (c *Client) Sync(ctx context.Context) (*proto.PaginatedWorldResponse, error) {
	return c.paginatedWorld(ctx)
}
