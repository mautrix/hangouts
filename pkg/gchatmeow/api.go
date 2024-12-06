package gchatmeow

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	pb "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"go.mau.fi/util/ptr"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func (c *Client) gcRequest(ctx context.Context, endpoint string, requestPB protoreflect.ProtoMessage, responsePB protoreflect.ProtoMessage) error {
	headers := http.Header{}
	if c.xsrfToken != "" {
		headers.Set("x-framework-xsrf-token", c.xsrfToken)
	}

	fmt.Printf("Sending Protocol Buffer request %s:\n%s\n", endpoint, requestPB)
	c.apiReqID++

	requestData, err := pb.Marshal(requestPB)
	if err != nil {
		return fmt.Errorf("failed to serialize protocol buffer: %v", err)
	}

	params := url.Values{}
	params.Set("c", strconv.FormatInt(c.apiReqID, 10))
	params.Set("rt", "b")
	res, err := c.baseRequest(
		ctx,
		fmt.Sprintf("%s/api/%s", gcBaseURL, endpoint),
		"application/x-protobuf",
		"proto",
		requestData,
		headers,
		params,
		http.MethodPost,
	)
	if err != nil {
		return err
	}

	if err := pb.Unmarshal(res, responsePB); err != nil {
		return fmt.Errorf("failed to decode Protocol Buffer response: %v", err)
	}

	fmt.Printf("Received Protocol Buffer response:\n%s\n", responsePB)
	return nil
}

func (c *Client) baseRequest(
	ctx context.Context,
	urlStr string,
	contentType string,
	responseType string,
	data []byte,
	headers http.Header,
	params url.Values,
	method string,
) ([]byte, error) {
	if headers == nil {
		headers = http.Header{}
	}

	if contentType != "" {
		headers.Set("content-type", contentType)
	}

	if responseType == "proto" {
		headers.Set("X-Goog-Encode-Response-If-Executable", "base64")
	}

	if params == nil {
		params = url.Values{}
	}
	params.Set("alt", responseType)
	params.Set("key", apiKey)

	res, err := c.session.Fetch(ctx, method, urlStr, params, headers, true, data)

	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (c *Client) getSelfUserStatus(ctx context.Context) (*proto.GetSelfUserStatusResponse, error) {
	request := &proto.GetSelfUserStatusRequest{
		RequestHeader: c.gcRequestHeader,
	}
	response := &proto.GetSelfUserStatusResponse{}
	err := c.gcRequest(ctx, "get_self_user_status", request, response)
	return response, err
}

func (c *Client) getMembers(ctx context.Context, gcid *string) (*proto.GetMembersResponse, error) {
	request := &proto.GetMembersRequest{
		RequestHeader: c.gcRequestHeader,
		MemberIds: []*proto.MemberId{
			&proto.MemberId{Id: &proto.MemberId_UserId{
				UserId: &proto.UserId{Id: gcid},
			}},
		},
	}
	response := &proto.GetMembersResponse{}
	err := c.gcRequest(ctx, "get_members", request, response)
	return response, err
}

func (c *Client) paginatedWorld(ctx context.Context) (*proto.PaginatedWorldResponse, error) {
	request := &proto.PaginatedWorldRequest{
		RequestHeader:       c.gcRequestHeader,
		FetchFromUserSpaces: ptr.Ptr(true),
		FetchOptions: []proto.PaginatedWorldRequest_FetchOptions{
			proto.PaginatedWorldRequest_EXCLUDE_GROUP_LITE,
		},
	}
	response := &proto.PaginatedWorldResponse{}
	err := c.gcRequest(ctx, "paginated_world", request, response)
	return response, err
}
