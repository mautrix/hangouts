package gchatmeow

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/cookies"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/endpoints"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/methods"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatproto"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/types"

	"github.com/rs/zerolog"
	"golang.org/x/net/html"
	"golang.org/x/net/proxy"
)

type EventHandler func(evt any)
type ClientOpts struct {
	GChatClientOpts *GChatClientOpts
	Cookies         *cookies.Cookies
	EventHandler    EventHandler
}

func (opts *ClientOpts) updateDefaultGChatClientOpts() {
	if opts.GChatClientOpts == nil {
		opts.GChatClientOpts = &GChatClientOpts{}
	}

	gchatOpts := opts.GChatClientOpts
	if gchatOpts.ClientType == gchatproto.ClientType_CLIENT_TYPE_UNKNOWN {
		gchatOpts.ClientType = gchatproto.ClientType_CLIENT_TYPE_WEB_DYNTO
	}

	if gchatOpts.ClientVersion == 0 {
		gchatOpts.ClientVersion = 1
	}

	if gchatOpts.Capabilities == nil {
		// 2 = fully supported, these are my default client capabilites I see
		gchatOpts.Capabilities = &gchatproto.ClientFeatureCapabilities{
			SpamRoomInvitesLevel: 2,
			TombstoneLevel:       2,
			ThreadedSpacesLevel:  2,
			FlatNamedRoomTopicOrderingByCreationTimeLevel: 2,
			TargetAudienceLevel:                           2,
			GroupScopedCapabilitiesLevel:                  2,
			RosterAsMemberSupportLevel:                    2,
			TombstoneInDmsAndUfrsLevel:                    2,
			QuotedMessageSupportLevel:                     2,
			RenderAnnouncementSpacesLevel:                 2,
			DarkLaunchSpaceSupport:                        2,
			AvoidHttp_400ErrorSupportLevel:                2,
			CustomHyperlinkLevel:                          2,
			SnippetsForNamedRooms:                         2,
			CanAddContinuousDirectAddGroups:               2,
			DriveSmartChipLevel:                           2,
			GsuiteIntegrationInNativeRendererLevel:        2,
			MentionsShortcutLevel:                         2,
			StarredShortcutLevel:                          2,
			SearchSnippetAndKeywordHighlightLevel:         2,
			CanHandleBatchReactionUpdate:                  2,
			LongerGroupSnippetsLevel:                      2,
			AddExistingAppsLevel:                          2,
		}
	}
}

type GChatClientOpts struct {
	ClientType    gchatproto.ClientType
	ClientVersion int64
	Locale        string
	Capabilities  *gchatproto.ClientFeatureCapabilities
}

type Client struct {
	Logger       zerolog.Logger
	cookies      *cookies.Cookies
	rc           *RealtimeClient
	http         *http.Client
	httpProxy    func(*http.Request) (*url.URL, error)
	socksProxy   proxy.Dialer
	eventHandler EventHandler
	XSRFToken    string
	XClientData  string
	opts         *ClientOpts
}

func NewClient(opts *ClientOpts, logger zerolog.Logger) *Client {
	cli := Client{
		http: &http.Client{
			Transport: &http.Transport{
				DialContext:           (&net.Dialer{Timeout: 10 * time.Second}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 40 * time.Second,
				ForceAttemptHTTP2:     true,
			},
			Timeout: 60 * time.Second,
		},
		Logger:      logger,
		opts:        opts,
		XClientData: "COSJywE=", // base64 encoded protobuf data
	}

	err := cli.setupClientOptions()
	if err != nil {
		log.Fatal(err)
	}

	cli.rc = cli.newRealtimeClient()

	return &cli
}

func (c *Client) setupClientOptions() error {
	if c.opts == nil {
		return fmt.Errorf("client options struct can not be nil")
	}

	if c.opts.EventHandler != nil {
		c.SetEventHandler(c.opts.EventHandler)
	}

	if c.opts.Cookies != nil {
		c.cookies = c.opts.Cookies
	} else {
		c.cookies = cookies.NewCookies()
	}

	// set/update default values for gchat client options
	c.opts.updateDefaultGChatClientOpts()

	return nil
}

func (c *Client) LoadMessagesPage() (*types.InitialConfigData, error) {
	headers := c.buildHeaders(types.HeaderOpts{
		WithCookies:     true,
		WithXClientData: true,
		Extra: map[string]string{
			"Referer": "https://mail.google.com/",
		},
	})

	url := fmt.Sprintf("%s?%s", endpoints.MOLE_BASE_URL, "wfi=gtn-brain-iframe-id&hs=%5B%22h_hs%22%2Cnull%2Cnull%2C%5B1%2C0%5D%2Cnull%2Cnull%2C%22gmail.pinto-server_20240610.06_p0%22%2C1%2Cnull%2C%5B15%2C48%2C6%2C43%2C36%2C35%2C26%2C44%2C39%2C46%2C41%2C18%2C24%2C11%2C21%2C14%2C1%2C51%2C53%5D%2Cnull%2Cnull%2C%22GgsaQvppfqU.en..es5%22%2Cnull%2Cnull%2Cnull%2C%5B2%5D%2Cnull%2C0%5D&hl=en&lts=chat%2Fhome&shell=9&has_stream_view=false&origin=https%3A%2F%2Fmail.google.com")
	resp, respBody, err := c.MakeRequest(url, http.MethodGet, headers, nil, types.NONE)
	if err != nil {
		return nil, err
	}

	c.cookies.UpdateFromResponse(resp)

	doc, err := html.Parse(bytes.NewReader(respBody))
	if err != nil {
		return nil, fmt.Errorf("failed to parse doc string for initial messaging page (%e)", err)
	}

	scriptTags := findScriptTags(doc)
	initialData, err := c.parseInitialMessagesHTML(scriptTags)
	if err != nil {
		return nil, err
	}

	c.XSRFToken = initialData.PageConfig.XSRFToken
	c.Logger.Info().Str("X-Framework-Xsrf-Token", c.XSRFToken).Msg("Successfully loaded initial messaging page config")

	//err = c.cookies.SaveToTxt()
	//if err != nil {
	//	log.Fatal(err)
	//}

	return initialData, nil
}

func (c *Client) Connect() error {
	return c.rc.Connect()
}

func (c *Client) Disconnect() error {
	return nil
}

func (c *Client) SetProxy(proxyAddr string) error {
	proxyParsed, err := url.Parse(proxyAddr)
	if err != nil {
		return err
	}

	if proxyParsed.Scheme == "http" || proxyParsed.Scheme == "https" {
		c.httpProxy = http.ProxyURL(proxyParsed)
		c.http.Transport.(*http.Transport).Proxy = c.httpProxy
	} else if proxyParsed.Scheme == "socks5" {
		c.socksProxy, err = proxy.FromURL(proxyParsed, &net.Dialer{Timeout: 20 * time.Second})
		if err != nil {
			return err
		}
		contextDialer, ok := c.socksProxy.(proxy.ContextDialer)
		if ok {
			c.http.Transport.(*http.Transport).DialContext = contextDialer.DialContext
		}
	}

	c.Logger.Debug().
		Str("scheme", proxyParsed.Scheme).
		Str("host", proxyParsed.Host).
		Msg("Using proxy")
	return nil
}

func (c *Client) SetEventHandler(handler EventHandler) {
	c.eventHandler = handler
}

func (c *Client) buildRequestHeader() *gchatproto.RequestHeader {
	return &gchatproto.RequestHeader{
		TraceId:                   methods.RandomInt64(),
		ClientType:                c.opts.GChatClientOpts.ClientType,
		ClientVersion:             c.opts.GChatClientOpts.ClientVersion,
		Locale:                    c.opts.GChatClientOpts.Locale,
		ClientFeatureCapabilities: c.opts.GChatClientOpts.Capabilities,
	}
}
