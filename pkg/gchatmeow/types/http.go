package types

type RequestMethod string

const (
	POST   RequestMethod = "POST"
	GET    RequestMethod = "GET"
	PUT    RequestMethod = "PUT"
	DELETE RequestMethod = "DELETE"
)

type ContentType string

const (
	NONE                     ContentType = ""
	JSON                     ContentType = "application/json"
	JSON_PLAINTEXT_UTF8      ContentType = "application/json; charset=UTF-8"
	JSON_LINKEDIN_NORMALIZED ContentType = "application/vnd.linkedin.normalized+json+2.1"
	FORM                     ContentType = "application/x-www-form-urlencoded"
	FORM_PLAINTEXT_UTF8      ContentType = "application/x-www-form-urlencoded;charset=UTF-8"
	GRAPHQL                  ContentType = "application/graphql"
	TEXT_EVENTSTREAM         ContentType = "text/event-stream"
	PLAINTEXT_UTF8           ContentType = "text/plain;charset=UTF-8"
	IMAGE_JPEG               ContentType = "image/jpeg"
	VIDEO_MP4                ContentType = "video/mp4"
	PROTOBUF                 ContentType = "application/x-protobuf"
)

type HeaderOpts struct {
	WithCookies     bool
	WithXSRFToken   bool
	WithXClientData bool
	Referer         string
	Origin          string
	Extra           map[string]string
}
