package connector

import (
	"google.golang.org/protobuf/encoding/prototext"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func (c *GChatClient) makePortalKey(evt *proto.Event) networkid.PortalKey {
	return networkid.PortalKey{
		ID:       networkid.PortalID(evt.GroupId.String()),
		Receiver: c.userLogin.ID,
	}
}

func (c *GChatClient) makeMessageId(portal *bridgev2.Portal, topicId, msgId string) *proto.MessageId {
	groupId := &proto.GroupId{}
	prototext.Unmarshal([]byte(portal.ID), groupId)
	return &proto.MessageId{
		ParentId: &proto.MessageParentId{
			Parent: &proto.MessageParentId_TopicId{
				TopicId: &proto.TopicId{
					GroupId: groupId,
					TopicId: topicId,
				},
			},
		},
		MessageId: msgId,
	}
}
