package query

import urlquery "github.com/google/go-querystring/query"

type UploadMediaQuery struct {
	GroupID         string `url:"group_id,omitempty"`
	TopicID         string `url:"topic_id,omitempty"`
	MessageID       string `url:"message_id,omitempty"`
	Otr             string `url:"otr,omitempty"`
	TranscodedVideo string `url:"transcoded_video,omitempty"`
	UploadType      string `url:"upload_type,omitempty"`
}

func (q *UploadMediaQuery) Encode() (string, error) {
	v, err := urlquery.Values(q)
	if err != nil {
		return "", err
	}
	return v.Encode(), nil
}
