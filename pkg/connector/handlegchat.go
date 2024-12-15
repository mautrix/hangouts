package connector

import (
	"context"
	"fmt"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/simplevent"
	bridgeEvt "maunium.net/go/mautrix/event"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
	"go.mau.fi/util/ptr"
)

func (c *GChatClient) onStreamEvent(ctx context.Context, raw any) {
	evt, ok := raw.(*proto.Event)
	if !ok {
		fmt.Println("Invalid event", raw)
		return
	}
	switch *evt.Type {
	case proto.Event_MESSAGE_POSTED:
		msg := evt.Body.GetMessagePosted().Message
		senderId := *msg.Creator.UserId.Id
		c.userLogin.Bridge.QueueRemoteEvent(c.userLogin, &simplevent.Message[*proto.Message]{
			EventMeta: simplevent.EventMeta{
				Type: bridgev2.RemoteEventMessage,
				PortalKey: networkid.PortalKey{
					ID:       networkid.PortalID(evt.GroupId.String()),
					Receiver: c.userLogin.ID,
				},
				// CreatePortal: true,
				Sender: bridgev2.EventSender{
					// IsFromMe:    isFromMe,
					SenderLogin: networkid.UserLoginID(senderId),
					Sender:      networkid.UserID(senderId),
				},
				// Timestamp: evtData.CreatedAt,
			},
			ID:   networkid.MessageID(*msg.LocalId),
			Data: msg,
			ConvertMessageFunc: func(ctx context.Context, portal *bridgev2.Portal, intent bridgev2.MatrixAPI, data *proto.Message) (*bridgev2.ConvertedMessage, error) {
				return c.convertToMatrix(ctx, portal, intent, data), nil
			},
		})
	}
}

func (c *GChatClient) convertToMatrix(ctx context.Context, portal *bridgev2.Portal, intent bridgev2.MatrixAPI, msg *proto.Message) *bridgev2.ConvertedMessage {
	parts := make([]*bridgev2.ConvertedMessagePart, 0)

	textPart := &bridgev2.ConvertedMessagePart{
		ID:   "",
		Type: bridgeEvt.EventMessage,
		Content: &bridgeEvt.MessageEventContent{
			MsgType: bridgeEvt.MsgText,
			Body:    *msg.TextBody,
		},
	}

	if len(textPart.Content.Body) > 0 {
		parts = append(parts, textPart)
	}

	for _, annotation := range msg.Annotations {
		attachmentPart, err := c.gcAnnotationToMatrix(ctx, portal, intent, annotation)
		if err != nil {
			fmt.Println(err)
			continue
		}
		parts = append(parts, attachmentPart)
	}

	cm := &bridgev2.ConvertedMessage{
		Parts: parts,
	}

	parentId := msg.Id.ParentId.GetTopicId().TopicId
	if parentId != nil {
		cm.ThreadRoot = ptr.Ptr(networkid.MessageID(*parentId))
	}
	if msg.ReplyTo != nil {
		replyTo := msg.ReplyTo.Id.MessageId
		if replyTo != nil {
			cm.ReplyTo = &networkid.MessageOptionalPartID{MessageID: networkid.MessageID(*replyTo)}
		}
	}

	// cm.MergeCaption()

	return cm
}
