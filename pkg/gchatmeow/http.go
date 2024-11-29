package gchatmeow

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"google.golang.org/protobuf/proto"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/endpoints"
	protoUtil "go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/types"
)

const MaxHTTPRetries = 5

var (
	ErrRedirectAttempted  = errors.New("redirect attempted")
	ErrTokenInvalidated   = errors.New("access token is no longer valid")
	ErrChallengeRequired  = errors.New("challenge required")
	ErrConsentRequired    = errors.New("consent required")
	ErrAccountSuspended   = errors.New("account suspended")
	ErrRequestFailed      = errors.New("failed to send request")
	ErrResponseReadFailed = errors.New("failed to read response body")
	ErrMaxRetriesReached  = errors.New("maximum retries reached")
)

func isPermanentRequestError(err error) bool {
	return errors.Is(err, ErrTokenInvalidated) ||
		errors.Is(err, ErrChallengeRequired) ||
		errors.Is(err, ErrConsentRequired) ||
		errors.Is(err, ErrAccountSuspended)
}

func (c *Client) makeAPIRequest(endpoint endpoints.APIEndpoint, payload proto.Message, response proto.Message, extraHeaders map[string]string) error {
	// rt = response_type=binary
	url := fmt.Sprintf("%s%s?c=1&rt=b", endpoints.API_BASE_URL, endpoint)

	headers := c.buildHeaders(types.HeaderOpts{
		WithCookies:     true,
		WithXSRFToken:   true,
		WithXClientData: true,
		Extra:           extraHeaders,
	})

	payloadBytes, err := protoUtil.EncodeProtoMessage(payload)
	if err != nil {
		return err
	}

	_, respBody, err := c.MakeRequest(url, http.MethodPost, headers, payloadBytes, types.PROTOBUF)
	if err != nil {
		return err
	}

	return protoUtil.DecodeProtoMessage(respBody, response)
}

func (c *Client) MakeRequest(url string, method string, headers http.Header, payload []byte, contentType types.ContentType) (*http.Response, []byte, error) {
	var attempts int
	for {
		attempts++
		start := time.Now()
		resp, respDat, err := c.makeRequestDirect(url, method, headers, payload, contentType)
		dur := time.Since(start)
		if err == nil {
			c.Logger.Debug().
				Str("url", url).
				Str("method", method).
				Dur("duration", dur).
				Msg("Request successful")
			return resp, respDat, nil
		} else if attempts > MaxHTTPRetries {
			c.Logger.Err(err).
				Str("url", url).
				Str("method", method).
				Dur("duration", dur).
				Msg("Request failed, giving up")
			return nil, nil, fmt.Errorf("%w: %w", ErrMaxRetriesReached, err)
		} else if isPermanentRequestError(err) {
			c.Logger.Err(err).
				Str("url", url).
				Str("method", method).
				Dur("duration", dur).
				Msg("Request failed, cannot be retried")
			return nil, nil, err
		} else if errors.Is(err, ErrRedirectAttempted) {
			location := resp.Header.Get("Location")
			c.Logger.Err(err).
				Str("url", url).
				Str("location", location).
				Str("method", method).
				Dur("duration", dur).
				Msg("Redirect attempted")
			return resp, nil, err
		}
		c.Logger.Err(err).
			Str("url", url).
			Str("method", method).
			Dur("duration", dur).
			Msg("Request failed, retrying")
		time.Sleep(time.Duration(attempts) * 3 * time.Second)
	}
}

func (c *Client) makeRequestDirect(url string, method string, headers http.Header, payload []byte, contentType types.ContentType) (*http.Response, []byte, error) {
	newRequest, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, nil, err
	}

	if contentType != types.NONE {
		headers.Set("content-type", string(contentType))
	}

	newRequest.Header = headers

	response, err := c.http.Do(newRequest)
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if err != nil {
		if errors.Is(err, ErrRedirectAttempted) {
			return response, nil, err
		}
		c.Logger.Warn().Str("error", err.Error()).Msg("Http request error")
		// c.UpdateProxy(fmt.Sprintf("http request error: %v", err.Error()))
		return nil, nil, fmt.Errorf("%w: %w", ErrRequestFailed, err)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %w", ErrResponseReadFailed, err)
	}

	return response, responseBody, nil
}
