package connector

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/simplevent"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
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
				Timestamp: time.UnixMicro(*msg.CreateTime),
			},
			ID:                 networkid.MessageID(*msg.Id.MessageId),
			Data:               msg,
			ConvertMessageFunc: c.msgConv.ToMatrix,
		})
	}

	c.setPortalRevision(ctx, evt)

	c.handleReaction(ctx, evt)
}

func (c *GChatClient) handleReaction(ctx context.Context, evt *proto.Event) {
	reaction := evt.Body.GetMessageReaction()
	if reaction == nil {
		return
	}

	var eventType bridgev2.RemoteEventType
	if reaction.GetType() == proto.MessageReactionEvent_ADD {
		eventType = bridgev2.RemoteEventReaction
	} else {
		eventType = bridgev2.RemoteEventReactionRemove

	}

	sender := reaction.UserId.GetId()
	messageId := reaction.MessageId.GetMessageId()
	c.userLogin.Bridge.QueueRemoteEvent(c.userLogin, &simplevent.Reaction{
		EventMeta: simplevent.EventMeta{
			Type: eventType,
			LogContext: func(c zerolog.Context) zerolog.Context {
				return c.
					Str("message_id", messageId).
					Str("sender", sender).
					Str("emoji", reaction.Emoji.GetUnicode()).
					Str("type", reaction.GetType().String())
			},
			PortalKey: c.makePortalKey(evt),
			Timestamp: time.UnixMicro(*reaction.Timestamp),
			Sender: bridgev2.EventSender{
				IsFromMe: sender == string(c.userLogin.ID),
				Sender:   networkid.UserID(sender),
			},
		},
		EmojiID:       networkid.EmojiID(reaction.Emoji.GetUnicode()),
		Emoji:         reaction.Emoji.GetUnicode(),
		TargetMessage: networkid.MessageID(messageId),
	})
}
