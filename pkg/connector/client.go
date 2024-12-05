package connector

import (
	"context"
	"fmt"
	"time"

	"go.mau.fi/util/ptr"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/simplevent"
	bridgeEvt "maunium.net/go/mautrix/event"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

type GChatClient struct {
	UserLogin *bridgev2.UserLogin
	Client    *gchatmeow.Client
}

var (
	_ bridgev2.NetworkAPI = (*GChatClient)(nil)
)

func (c *GChatClient) Connect(ctx context.Context) error {
	c.Client.OnConnect.AddObserver(func(interface{}) { c.onConnect(ctx) })
	c.Client.OnStreamEvent.AddObserver(func(evt interface{}) { c.onStreamEvent(ctx, evt) })
	return c.Client.Connect(ctx, time.Duration(90)*time.Minute)
}

func (c *GChatClient) Disconnect() {
}

func (c *GChatClient) GetCapabilities(ctx context.Context, portal *bridgev2.Portal) *bridgev2.NetworkRoomCapabilities {
	return nil
}

func (c *GChatClient) GetChatInfo(ctx context.Context, portal *bridgev2.Portal) (*bridgev2.ChatInfo, error) {
	return nil, nil
}

func (c *GChatClient) GetUserInfo(ctx context.Context, ghost *bridgev2.Ghost) (*bridgev2.UserInfo, error) {
	return nil, nil
}

func (c *GChatClient) IsLoggedIn() bool {
	return false
}

func (c *GChatClient) IsThisUser(ctx context.Context, userID networkid.UserID) bool {
	return networkid.UserID(c.UserLogin.ID) == userID
}

func (c *GChatClient) LogoutRemote(ctx context.Context) {
}

func (c *GChatClient) onConnect(ctx context.Context) {
	res, err := c.Client.Sync(ctx)
	if err != nil {
		fmt.Println((err))
		return
	}
	for _, item := range res.WorldItems {
		name := item.RoomName
		if name == nil {
			name = ptr.Ptr("dm")
		}
		memberMap := map[networkid.UserID]bridgev2.ChatMember{}
		memberMap[networkid.UserID(c.UserLogin.ID)] = bridgev2.ChatMember{
			EventSender: bridgev2.EventSender{
				IsFromMe: true,
				Sender:   networkid.UserID(c.UserLogin.ID),
			},
		}
		c.UserLogin.Bridge.QueueRemoteEvent(c.UserLogin, &simplevent.ChatResync{
			EventMeta: simplevent.EventMeta{
				Type: bridgev2.RemoteEventChatResync,
				PortalKey: networkid.PortalKey{
					ID:       networkid.PortalID(item.GroupId.String()),
					Receiver: c.UserLogin.ID,
				},
				CreatePortal: true,
			},
			ChatInfo: &bridgev2.ChatInfo{
				Name: name,
				Members: &bridgev2.ChatMemberList{
					MemberMap: memberMap,
				},
			},
		})

	}
}

func (c *GChatClient) onStreamEvent(ctx context.Context, raw any) {
	evt, ok := raw.(*proto.Event)
	if !ok {
		fmt.Println("Invalid event", raw)
		return
	}
	fmt.Println("- onEvent", evt)
	switch *evt.Type {
	case proto.Event_MESSAGE_POSTED:
		msg := evt.Body.GetMessagePosted().Message
		senderId := *msg.Creator.UserId.Id
		c.UserLogin.Bridge.QueueRemoteEvent(c.UserLogin, &simplevent.Message[*proto.Message]{
			EventMeta: simplevent.EventMeta{
				Type: bridgev2.RemoteEventMessage,
				// LogContext: func(c zerolog.Context) zerolog.Context {
				// 	return c.
				// 		Str("message_id", evtData.MessageID).
				// 		Str("sender", sender.IDStr).
				// 		Str("sender_login", sender.ScreenName).
				// 		Bool("is_from_me", isFromMe)
				// },
				PortalKey: networkid.PortalKey{
					ID:       networkid.PortalID(evt.GroupId.String()),
					Receiver: c.UserLogin.ID,
				},
				// CreatePortal: true,
				Sender: bridgev2.EventSender{
					// IsFromMe:    isFromMe,
					SenderLogin: networkid.UserLoginID(senderId),
					Sender:      networkid.UserID(senderId),
				},
				// Timestamp: evtData.CreatedAt,
			},
			ID: networkid.MessageID(*msg.LocalId),
			// TargetMessage: networkid.MessageID(evtData.MessageID),
			// Data:          XMDFromEventMessage(&evtData),
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

	cm := &bridgev2.ConvertedMessage{
		Parts: parts,
	}

	return cm
}
