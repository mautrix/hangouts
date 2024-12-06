package gchatmeow

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"time"
)

const (
	connectTimeout   = 30 * time.Second
	requestTimeout   = 30 * time.Second
	maxRetries       = 3
	originURL        = "https://chat.google.com"
	latestChromeVer  = "114"
	latestFirefoxVer = "114"
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
)

var (
	chromeVersionRegex  = regexp.MustCompile(`Chrome/\d+\.\d+\.\d+\.\d+`)
	firefoxVersionRegex = regexp.MustCompile(`Firefox/\d+.\d+`)
)

type FetchResponse struct {
	Code    int
	Headers http.Header
	Body    []byte
}

type Session struct {
	client    *http.Client
	proxy     *url.URL
	userAgent string
	cookies   *cookiejar.Jar
}

type NetworkError struct {
	message string
}

func (e *NetworkError) Error() string {
	return e.message
}

func NewSession(cookies *Cookies, proxyURL string, userAgent string) (*Session, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %v", err)
	}

	var proxy *url.URL
	if proxyURL != "" {
		proxy, err = url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %v", err)
		}
	}

	if userAgent != "" {
		userAgent = chromeVersionRegex.ReplaceAllString(userAgent, fmt.Sprintf("Chrome/%s.0.0.0", latestChromeVer))
		userAgent = firefoxVersionRegex.ReplaceAllString(userAgent, fmt.Sprintf("Firefox/%s.0", latestFirefoxVer))
	} else {
		userAgent = defaultUserAgent
	}

	googleURL, _ := url.Parse("https://chat.google.com")
	var googleCookies []*http.Cookie
	for name, value := range map[string]string{
		"COMPASS": cookies.COMPASS,
		"SSID":    cookies.SSID,
		"SID":     cookies.SID,
		"OSID":    cookies.OSID,
		"HSID":    cookies.HSID,
	} {
		googleCookies = append(googleCookies, &http.Cookie{
			Name:   name,
			Value:  value,
			Domain: "chat.google.com",
			Path:   "/",
		})
	}
	jar.SetCookies(googleURL, googleCookies)

	client := &http.Client{
		Jar: jar,
		Transport: &http.Transport{
			Proxy:              http.ProxyURL(proxy),
			TLSClientConfig:    nil, // equivalent to ssl=False in Python
			DisableCompression: true,
		},
		Timeout: connectTimeout,
	}

	return &Session{
		client:    client,
		proxy:     proxy,
		userAgent: userAgent,
		cookies:   jar,
	}, nil
}

func (s *Session) Fetch(ctx context.Context, method, urlStr string, params url.Values, headers http.Header, allowRedirects bool, data []byte) (*FetchResponse, error) {
	var lastErr error

	for retry := 0; retry < maxRetries; retry++ {
		resp, err := s.FetchRaw(ctx, method, urlStr, params, headers, allowRedirects, data)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = &NetworkError{fmt.Sprintf("unexpected status: %d %s", resp.StatusCode, resp.Status)}
			continue
		}

		return &FetchResponse{
			Code:    resp.StatusCode,
			Headers: resp.Header,
			Body:    body,
		}, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %v", maxRetries, lastErr)
}

func (s *Session) FetchRaw(ctx context.Context, method, urlStr string, params url.Values, headers http.Header, allowRedirects bool, data []byte) (*http.Response, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	if !regexp.MustCompile(`\.google\.com$`).MatchString(parsedURL.Host) {
		return nil, fmt.Errorf("expected google.com domain")
	}

	if params != nil {
		parsedURL.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if headers != nil {
		req.Header = headers
	}
	req.Header.Set("User-Agent", s.userAgent)
	req.Header.Set("Connection", "Keep-Alive")

	client := s.client
	if !allowRedirects {
		client = &http.Client{
			Jar:       s.cookies,
			Transport: s.client.Transport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: s.client.Timeout,
		}
	}

	return client.Do(req)
}
