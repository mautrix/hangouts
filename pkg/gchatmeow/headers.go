package gchatmeow

import (
	"net/http"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/types"
)

const BrowserName = "Chrome"
const ChromeVersion = "118"
const ChromeVersionFull = ChromeVersion + ".0.5993.89"
const UserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + ChromeVersion + ".0.0.0 Safari/537.36"
const SecCHUserAgent = `"Chromium";v="` + ChromeVersion + `", "Google Chrome";v="` + ChromeVersion + `", "Not-A.Brand";v="99"`
const SecCHFullVersionList = `"Chromium";v="` + ChromeVersionFull + `", "Google Chrome";v="` + ChromeVersionFull + `", "Not-A.Brand";v="99.0.0.0"`
const OSName = "Linux"
const OSVersion = "6.5.0"
const SecCHPlatform = `"` + OSName + `"`
const SecCHPlatformVersion = `"` + OSVersion + `"`
const SecCHMobile = "?0"
const SecCHModel = ""
const SecCHPrefersColorScheme = "light"

var defaultConstantHeaders = http.Header{
	"accept":                      []string{"text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
	"accept-language":             []string{"en-US,en;q=0.9"},
	"user-agent":                  []string{UserAgent},
	"sec-ch-ua":                   []string{SecCHUserAgent},
	"sec-ch-ua-platform":          []string{SecCHPlatform},
	"sec-ch-prefers-color-scheme": []string{SecCHPrefersColorScheme},
	"sec-ch-ua-full-version-list": []string{SecCHFullVersionList},
	"sec-ch-ua-mobile":            []string{SecCHMobile},
	"sec-ch-ua-model":             []string{SecCHModel},
	"sec-ch-ua-platform-version":  []string{SecCHPlatformVersion},
}

func (c *Client) buildHeaders(opts types.HeaderOpts) http.Header {
	if opts.Extra == nil {
		opts.Extra = make(map[string]string, 0)
	}

	headers := defaultConstantHeaders.Clone()
	if opts.WithCookies {
		opts.Extra["cookie"] = c.cookies.String()
	}

	//if opts.ExpectProtoResponse {
	//	opts.Extra["X-Goog-Encode-Response-If-Executable"] = "base64"
	//}

	if opts.WithXSRFToken {
		opts.Extra["X-Framework-Xsrf-Token"] = c.XSRFToken
	}

	if opts.WithXClientData {
		opts.Extra["X-Client-Data"] = c.XClientData
	}

	if opts.Origin != "" {
		opts.Extra["origin"] = opts.Origin
	}

	if opts.Referer != "" {
		opts.Extra["referer"] = opts.Referer
	}

	for k, v := range opts.Extra {
		headers.Set(k, v)
	}

	return headers
}
