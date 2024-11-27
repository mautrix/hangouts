package cookies

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type GChatCookieName string

const (
	GMAIL_AT           GChatCookieName = "GMAIL_AT"
	COMPASS            GChatCookieName = "COMPASS"
	AEC                GChatCookieName = "AEC"
	SOCS               GChatCookieName = "SOCS"
	SID                GChatCookieName = "SID"
	Secure_1PSID       GChatCookieName = "__Secure-1PSID"
	Secure_3PSID       GChatCookieName = "__Secure-3PSID"
	HSID               GChatCookieName = "HSID"
	SSID               GChatCookieName = "SSID"
	APISID             GChatCookieName = "APISID"
	SAPISID            GChatCookieName = "SAPISID"
	Secure_1PAPISID    GChatCookieName = "__Secure-1PAPISID"
	Secure_3PAPISID    GChatCookieName = "__Secure-3PAPISID"
	Secure_1PSIDTS     GChatCookieName = "__Secure-1PSIDTS"
	Secure_3PSIDTS     GChatCookieName = "__Secure-3PSIDTS"
	OSID               GChatCookieName = "OSID"
	Secure_OSID        GChatCookieName = "__Secure-OSID"
	Host_GMAIL_SCH_GMN GChatCookieName = "__Host-GMAIL_SCH_GMN"
	Host_GMAIL_SCH_GMS GChatCookieName = "__Host-GMAIL_SCH_GMS"
	Host_GMAIL_SCH_GML GChatCookieName = "__Host-GMAIL_SCH_GML"
	SEARCH_SAMESITE    GChatCookieName = "SEARCH_SAMESITE"
	NID                GChatCookieName = "NID"
	Host_GMAIL_SCH     GChatCookieName = "__Host-GMAIL_SCH"
	SIDCC              GChatCookieName = "SIDCC"
	Secure_1PSIDCC     GChatCookieName = "__Secure-1PSIDCC"
	Secure_3PSIDCC     GChatCookieName = "__Secure-3PSIDCC"
)

type Cookies struct {
	store map[GChatCookieName]string
	lock  sync.RWMutex
}

func NewCookies() *Cookies {
	return &Cookies{
		store: make(map[GChatCookieName]string),
		lock:  sync.RWMutex{},
	}
}

func NewCookiesFromString(cookieStr string) *Cookies {
	c := NewCookies()
	cookieStrings := strings.Split(cookieStr, ";")
	fakeHeader := http.Header{}
	for _, cookieStr := range cookieStrings {
		trimmedCookieStr := strings.TrimSpace(cookieStr)
		if trimmedCookieStr != "" {
			fakeHeader.Add("Set-Cookie", trimmedCookieStr)
		}
	}
	fakeResponse := &http.Response{Header: fakeHeader}

	for _, cookie := range fakeResponse.Cookies() {
		c.store[GChatCookieName(cookie.Name)] = cookie.Value
	}
	return c
}

func (c *Cookies) String() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	var out []string
	for k, v := range c.store {
		out = append(out, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(out, "; ")
}

func (c *Cookies) IsCookieEmpty(key GChatCookieName) bool {
	return c.Get(key) == ""
}

func (c *Cookies) Get(key GChatCookieName) string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.store[key]
}

func (c *Cookies) Set(key GChatCookieName, value string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.store[key] = value
}

func (c *Cookies) UpdateFromResponse(r *http.Response) {
	c.lock.Lock()
	defer c.lock.Unlock()
	for _, cookie := range r.Cookies() {
		if cookie.MaxAge == 0 || cookie.Expires.Before(time.Now()) {
			delete(c.store, GChatCookieName(cookie.Name))
		} else {
			//log.Println(fmt.Sprintf("updated cookie %s to value %s", cookie.Name, cookie.Value))
			c.store[GChatCookieName(cookie.Name)] = cookie.Value
		}
	}
}

// this function is going to be removed. just here while im testing stuff, however we will need to update the cookies in db after every update
func (c *Cookies) SaveToTxt() error {
	cookieStr := c.String()
	return os.WriteFile("cookies.txt", []byte(cookieStr), os.ModePerm)
}
