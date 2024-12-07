package connector

import (
	"context"

	"google.golang.org/protobuf/encoding/prototext"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func portalToGroupId(portal *bridgev2.Portal) (*proto.GroupId, error) {
	groupId := &proto.GroupId{}
	err := prototext.Unmarshal([]byte(portal.ID), groupId)
	if err != nil {
		return nil, err
	}

	return groupId, nil
}

func (c *GChatClient) HandleMatrixMessage(ctx context.Context, msg *bridgev2.MatrixMessage) (message *bridgev2.MatrixMessageResponse, err error) {
	groupId, err := portalToGroupId(msg.Portal)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.CreateTopic(ctx, &proto.CreateTopicRequest{
		GroupId:  groupId,
		TextBody: &msg.Content.Body,
	})
	if err != nil {
		return nil, err
	}
	msgID := *res.Topic.Id.TopicId
	msg.AddPendingToIgnore(networkid.TransactionID(msgID))
	return &bridgev2.MatrixMessageResponse{
		DB: &database.Message{
			ID: networkid.MessageID(msgID),
		},
		RemovePending: networkid.TransactionID(msgID),
	}, nil
}
