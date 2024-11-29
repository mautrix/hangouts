package connector

import (
	"context"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/simplevent"
)

type GChatClient struct {
	UserLogin *bridgev2.UserLogin
	Client    *gchatmeow.Client
}

var (
	_ bridgev2.NetworkAPI = (*GChatClient)(nil)
)

func (c *GChatClient) Connect(ctx context.Context) error {
	_, err := c.Client.LoadMessagesPage()
	if err != nil {
		return err
	}
	err = c.Client.Connect()
	if err != nil {
		return err
	}
	return c.onConnect(ctx)
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

func (c *GChatClient) onConnect(ctx context.Context) error {
	res, err := c.Client.GetPaginatedWorlds(nil)
	if err != nil {
		return err
	}
	for _, item := range res.WorldItems {
		// TODO room name for DM, and full members list
		name := item.GetRoomName()
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
				Name: &name,
				Members: &bridgev2.ChatMemberList{
					MemberMap: memberMap,
				},
			},
		})
	}
	return nil
}
