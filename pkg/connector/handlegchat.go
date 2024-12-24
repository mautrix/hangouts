package connector

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/simplevent"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func (c *GChatClient) makeEventMeta(evt *proto.Event, typ bridgev2.RemoteEventType, senderId string, ts int64) simplevent.EventMeta {
	return simplevent.EventMeta{
		Type:      typ,
		PortalKey: c.makePortalKey(evt),
		Sender: bridgev2.EventSender{
			IsFromMe:    senderId == string(c.userLogin.ID),
			SenderLogin: networkid.UserLoginID(senderId),
			Sender:      networkid.UserID(senderId),
		},
		Timestamp: time.UnixMicro(ts),
	}
}

func (c *GChatClient) onStreamEvent(ctx context.Context, raw any) {
	evt, ok := raw.(*proto.Event)
	if !ok {
		fmt.Println("Invalid event", raw)
		return
	}

	switch evt.Type {
	case proto.Event_MESSAGE_POSTED:
		msg := evt.Body.GetMessagePosted().Message
		c.userLogin.Bridge.QueueRemoteEvent(c.userLogin, &simplevent.Message[*proto.Message]{
			EventMeta:          c.makeEventMeta(evt, bridgev2.RemoteEventMessage, msg.Creator.UserId.Id, msg.CreateTime),
			ID:                 networkid.MessageID(msg.Id.MessageId),
			Data:               msg,
			ConvertMessageFunc: c.msgConv.ToMatrix,
		})
	case proto.Event_MESSAGE_UPDATED:
		msg := evt.Body.GetMessagePosted().Message
		eventMeta := c.makeEventMeta(evt, bridgev2.RemoteEventEdit, msg.Creator.UserId.Id, msg.LastEditTime)
		c.userLogin.Bridge.QueueRemoteEvent(c.userLogin, &simplevent.Message[*proto.Message]{
			EventMeta:       eventMeta,
			ID:              networkid.MessageID(msg.Id.MessageId),
			TargetMessage:   networkid.MessageID(msg.Id.MessageId),
			Data:            msg,
			ConvertEditFunc: c.ConvertEdit,
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
	eventMeta := c.makeEventMeta(evt, eventType, sender, reaction.Timestamp)
	eventMeta.LogContext = func(c zerolog.Context) zerolog.Context {
		return c.
			Str("message_id", messageId).
			Str("sender", sender).
			Str("emoji", reaction.Emoji.GetUnicode()).
			Str("type", reaction.GetType().String())
	}
	c.userLogin.Bridge.QueueRemoteEvent(c.userLogin, &simplevent.Reaction{
		EventMeta:     eventMeta,
		EmojiID:       networkid.EmojiID(reaction.Emoji.GetUnicode()),
		Emoji:         reaction.Emoji.GetUnicode(),
		TargetMessage: networkid.MessageID(messageId),
	})
}

func (c *GChatClient) ConvertEdit(ctx context.Context, portal *bridgev2.Portal, intent bridgev2.MatrixAPI, existing []*database.Message, msg *proto.Message) (*bridgev2.ConvertedEdit, error) {
	cm, err := c.msgConv.ToMatrix(ctx, portal, intent, msg)
	if err != nil {
		return nil, err
	}

	editPart := cm.Parts[len(cm.Parts)-1].ToEditPart(existing[len(existing)-1])
	editPart.Part.EditCount++

	return &bridgev2.ConvertedEdit{
		ModifiedParts: []*bridgev2.ConvertedEditPart{editPart},
	}, nil
}
