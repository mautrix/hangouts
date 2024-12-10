package connector

import (
	"context"
	"fmt"
	"time"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/simplevent"
	bridgeEvt "maunium.net/go/mautrix/event"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

type GChatClient struct {
	userLogin *bridgev2.UserLogin
	client    *gchatmeow.Client
	users     map[string]*proto.User
}

var (
	_ bridgev2.NetworkAPI = (*GChatClient)(nil)
)

func NewClient(userLogin *bridgev2.UserLogin, client *gchatmeow.Client) *GChatClient {
	return &GChatClient{
		userLogin: userLogin,
		client:    client,
		users:     map[string]*proto.User{},
	}
}

func (c *GChatClient) Connect(ctx context.Context) error {
	c.client.OnConnect.AddObserver(func(interface{}) { c.onConnect(ctx) })
	c.client.OnStreamEvent.AddObserver(func(evt interface{}) { c.onStreamEvent(ctx, evt) })
	return c.client.Connect(ctx, time.Duration(90)*time.Minute)
}

func (c *GChatClient) Disconnect() {
}

func (c *GChatClient) GetCapabilities(ctx context.Context, portal *bridgev2.Portal) *bridgev2.NetworkRoomCapabilities {
	return &bridgev2.NetworkRoomCapabilities{}
}

func (c *GChatClient) GetChatInfo(ctx context.Context, portal *bridgev2.Portal) (*bridgev2.ChatInfo, error) {
	return nil, nil
}

func (c *GChatClient) GetUserInfo(ctx context.Context, ghost *bridgev2.Ghost) (*bridgev2.UserInfo, error) {
	return nil, nil
}

func (c *GChatClient) IsLoggedIn() bool {
	return true
}

func (c *GChatClient) IsThisUser(ctx context.Context, userID networkid.UserID) bool {
	return networkid.UserID(c.userLogin.ID) == userID
}

func (c *GChatClient) LogoutRemote(ctx context.Context) {
}

func (c *GChatClient) getUsers(ctx context.Context, userIds []*string) error {
	res, err := c.client.GetMembers(ctx, userIds)
	if err != nil {
		return err
	}
	for _, member := range res.Members {
		user := member.GetUser()
		c.users[*user.UserId.Id] = user
	}
	return nil
}

func (c *GChatClient) onConnect(ctx context.Context) {
	res, err := c.client.Sync(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	userIdMap := make(map[*string]struct{})
	for _, item := range res.WorldItems {
		if item.DmMembers != nil {
			for _, member := range item.DmMembers.Members {
				userIdMap[member.Id] = struct{}{}
			}
		}
	}
	userIds := make([]*string, len(userIdMap))
	i := 0
	for userId := range userIdMap {
		userIds[i] = userId
		i++
	}

	err = c.getUsers(ctx, userIds)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, item := range res.WorldItems {
		name := item.RoomName
		var gcMembers []*proto.UserId
		if item.DmMembers != nil {
			gcMembers = item.DmMembers.Members
			for _, member := range item.DmMembers.Members {
				if *member.Id != string(c.userLogin.ID) {
					name = c.users[*member.Id].Name
					break
				}
			}
		} else {
			group, err := c.client.GetGroup(ctx, &proto.GetGroupRequest{
				GroupId: item.GroupId,
				FetchOptions: []proto.GetGroupRequest_FetchOptions{
					proto.GetGroupRequest_MEMBERS,
				},
			})
			if err != nil {
				fmt.Println(err)
				continue
			}
			gcMembers = make([]*proto.UserId, len(group.Memberships))
			for i, membership := range group.Memberships {
				gcMembers[i] = membership.Id.MemberId.GetUserId()
			}

		}

		c.userLogin.Bridge.QueueRemoteEvent(c.userLogin, &simplevent.ChatResync{
			EventMeta: simplevent.EventMeta{
				Type: bridgev2.RemoteEventChatResync,
				PortalKey: networkid.PortalKey{
					ID:       networkid.PortalID(item.GroupId.String()),
					Receiver: c.userLogin.ID,
				},
				CreatePortal: true,
			},
			ChatInfo: &bridgev2.ChatInfo{
				Name:    name,
				Members: c.gcMembersToMatrix(gcMembers),
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
	cm.MergeCaption()

	return cm
}
