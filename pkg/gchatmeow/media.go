package gchatmeow

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/endpoints"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/query"
	protoUtil "go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatproto"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/types"
)

func (c *Client) UploadMedia(fileName string, mediaContentType string, mediaBytes []byte, uploadMediaQuery query.UploadMediaQuery) (*gchatproto.UploadMetadata, error) {
	extraHeaders := map[string]string{
		"chat-filename":                     fileName,
		"X-Goog-Upload-Command":             "start",
		"X-Goog-Upload-Content-Length":      strconv.Itoa(len(mediaBytes)),
		"X-Goog-Upload-File-Name":           fileName,
		"X-Goog-Upload-Header-Content-Type": mediaContentType,
		"X-Goog-Upload-Protocol":            "resumable",
	}

	queryString, err := uploadMediaQuery.Encode()
	if err != nil {
		return nil, err
	}

	headers := c.buildHeaders(types.HeaderOpts{
		WithCookies:     true,
		WithXClientData: true,
		Extra:           extraHeaders,
	})
	url := fmt.Sprintf("%s?%s", endpoints.UPLOADS, queryString)
	resp, _, err := c.MakeRequest(url, http.MethodPost, headers, nil, types.FORM_PLAINTEXT_UTF8)
	if err != nil {
		return nil, err
	}

	respHeaders := resp.Header

	//uploadStatus := respHeaders.Get("X-Goog-Upload-Status")
	//uploadId := respHeaders.Get("X-GUploader-UploadID")
	uploadUrl := respHeaders.Get("X-Goog-Upload-URL")
	return c.finalizeMediaUpload(uploadUrl, mediaBytes, headers)
}

func (c *Client) finalizeMediaUpload(uploadUrl string, mediaBytes []byte, headers http.Header) (*gchatproto.UploadMetadata, error) {
	md5Hash := md5.Sum(mediaBytes)
	md5Entity := hex.EncodeToString(md5Hash[:])

	headers.Set("X-Goog-Upload-Command", "upload, finalize")
	headers.Set("X-Goog-Upload-Entity-MD5", md5Entity)
	headers.Set("X-Goog-Upload-Offset", "0")

	_, respBody, err := c.MakeRequest(uploadUrl, http.MethodPut, headers, mediaBytes, types.FORM_PLAINTEXT_UTF8)
	if err != nil {
		return nil, err
	}

	protoBytes, err := base64.StdEncoding.DecodeString(string(respBody))
	if err != nil {
		return nil, err
	}

	respData := &gchatproto.UploadMetadata{}
	return respData, protoUtil.DecodeProtoMessage(protoBytes, respData)
}
