package gchatmeow

import (
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/endpoints"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatproto"

	"go.mau.fi/util/random"
)

func (c *Client) GetPaginatedWorlds(worldSectionRequests []*gchatproto.WorldSectionRequest) (*gchatproto.PaginatedWorldResponse, error) {
	requestPayload := &gchatproto.PaginatedWorldRequest{
		RequestHeader:        c.buildRequestHeader(),
		WorldSectionRequests: worldSectionRequests,
	}

	response := &gchatproto.PaginatedWorldResponse{}
	return response, c.makeAPIRequest(endpoints.PAGINATED_WORLD, requestPayload, response, nil)
}

func (c *Client) ListMembers(payload *gchatproto.ListMembersRequest) (*gchatproto.ListMembersResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	extraHeaders := grabGoogChatSpaceId(payload.GroupId)
	if payload.SpaceId != nil {
		extraHeaders = map[string]string{
			"x-goog-chat-space-id": payload.GetSpaceId().SpaceId,
		}
	}

	response := &gchatproto.ListMembersResponse{}
	return response, c.makeAPIRequest(endpoints.LIST_MEMBERS, payload, response, extraHeaders)
}

func (c *Client) SendMessage(payload *gchatproto.CreateTopicRequest) (*gchatproto.CreateTopicResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	if payload.TopicAndMessageId == "" {
		payload.TopicAndMessageId = random.String(11)
	}

	extraHeaders := grabGoogChatSpaceId(payload.GroupId)
	response := &gchatproto.CreateTopicResponse{}
	return response, c.makeAPIRequest(endpoints.CREATE_TOPIC, payload, response, extraHeaders)
}

func (c *Client) ListMessages(payload *gchatproto.ListTopicsRequest) (*gchatproto.ListTopicsResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	extraHeaders := grabGoogChatSpaceId(payload.GroupId)
	response := &gchatproto.ListTopicsResponse{}
	return response, c.makeAPIRequest(endpoints.LIST_TOPICS, payload, response, extraHeaders)
}

func (c *Client) UpdateReaction(payload *gchatproto.UpdateReactionRequest) (*gchatproto.UpdateReactionResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.UpdateReactionResponse{}
	return response, c.makeAPIRequest(endpoints.UPDATE_REACTION, payload, response, nil)
}

func (c *Client) DeleteMessage(payload *gchatproto.DeleteMessageRequest) (*gchatproto.DeleteMessageResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.DeleteMessageResponse{}
	return response, c.makeAPIRequest(endpoints.DELETE_MESSAGE, payload, response, nil)
}

func (c *Client) EditMessage(payload *gchatproto.EditMessageRequest) (*gchatproto.EditMessageResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.EditMessageResponse{}
	return response, c.makeAPIRequest(endpoints.EDIT_MESSAGE, payload, response, nil)
}

func (c *Client) CreateGroup(payload *gchatproto.CreateGroupRequest) (*gchatproto.CreateGroupResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.CreateGroupResponse{}
	return response, c.makeAPIRequest(endpoints.CREATE_GROUP, payload, response, nil)
}

func (c *Client) AddMember(payload *gchatproto.CreateMembershipRequest) (*gchatproto.CreateMembershipResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.CreateMembershipResponse{}
	return response, c.makeAPIRequest(endpoints.CREATE_MEMBERSHIP, payload, response, nil)
}

func (c *Client) GetGroup(payload *gchatproto.GetGroupRequest) (*gchatproto.GetGroupResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.GetGroupResponse{}
	return response, c.makeAPIRequest(endpoints.GET_GROUP, payload, response, nil)
}

func (c *Client) SetMarkAsUnreadTimestamp(payload *gchatproto.SetMarkAsUnreadTimestampRequest) (*gchatproto.SetMarkAsUnreadTimestampResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.SetMarkAsUnreadTimestampResponse{}
	return response, c.makeAPIRequest(endpoints.MARK_AS_UNREAD, payload, response, nil)
}

func (c *Client) CreateDMExtended(payload *gchatproto.CreateDmExtendedRequest) (*gchatproto.CreateDmExtendedResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.CreateDmExtendedResponse{}
	return response, c.makeAPIRequest(endpoints.CREATE_DM_EXTENDED, payload, response, nil)
}

// this is also used to leave a group chat
func (c *Client) RemoveMemberships(payload *gchatproto.RemoveMembershipsRequest) (*gchatproto.RemoveMembershipsResponse, error) {
	if payload.RequestHeader == nil {
		payload.RequestHeader = c.buildRequestHeader()
	}

	response := &gchatproto.RemoveMembershipsResponse{}
	return response, c.makeAPIRequest(endpoints.REMOVE_MEMBERSHIPS, payload, response, nil)
}

func grabGoogChatSpaceId(groupId *gchatproto.GroupId) map[string]string {
	var xGoogleChatSpaceID string
	var extraHeaders map[string]string

	switch groupId.Id.(type) {
	case *gchatproto.GroupId_SpaceId:
		xGoogleChatSpaceID = groupId.GetSpaceId().GetSpaceId()
	}

	if xGoogleChatSpaceID != "" {
		extraHeaders = map[string]string{
			"x-goog-chat-space-id": xGoogleChatSpaceID,
		}
	}

	return extraHeaders
}
