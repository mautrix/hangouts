package connector

import (
	"context"
	"fmt"
	"strings"
	"time"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"
	"maunium.net/go/mautrix/bridgev2/simplevent"

	"go.mau.fi/util/ptr"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
	"go.mau.fi/mautrix-googlechat/pkg/msgconv"
)

type GChatClient struct {
	userLogin *bridgev2.UserLogin
	client    *gchatmeow.Client
	users     map[string]*proto.User
	msgConv   *msgconv.MessageConverter
}

var (
	_ bridgev2.NetworkAPI = (*GChatClient)(nil)
)

func NewClient(userLogin *bridgev2.UserLogin, client *gchatmeow.Client) *GChatClient {
	return &GChatClient{
		userLogin: userLogin,
		client:    client,
		users:     map[string]*proto.User{},
		msgConv:   msgconv.NewMessageConverter(client),
	}
}

func (c *GChatClient) Connect(ctx context.Context) error {
	c.client.OnConnect.AddObserver(func(interface{}) { c.onConnect(ctx) })
	c.client.OnStreamEvent.AddObserver(func(evt interface{}) { c.onStreamEvent(ctx, evt) })
	return c.client.Connect(ctx, time.Duration(90)*time.Minute)
}

func (c *GChatClient) Disconnect() {
}

var dmCaps = &bridgev2.NetworkRoomCapabilities{
	Replies: true,
}

var spaceCaps *bridgev2.NetworkRoomCapabilities

func init() {
	spaceCaps = ptr.Clone(dmCaps)
	spaceCaps.Threads = true
}

func (c *GChatClient) GetCapabilities(ctx context.Context, portal *bridgev2.Portal) *bridgev2.NetworkRoomCapabilities {
	if strings.Contains(string(portal.ID), "space") {
		return spaceCaps
	}
	return dmCaps
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

func (c *GChatClient) getUsers(ctx context.Context, userIds []string) error {
	res, err := c.client.GetMembers(ctx, userIds)
	if err != nil {
		return err
	}
	for _, member := range res.Members {
		user := member.GetUser()
		c.users[user.UserId.Id] = user
	}
	return nil
}

func (c *GChatClient) onConnect(ctx context.Context) {
	res, err := c.client.Sync(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	userIdMap := make(map[string]struct{})
	for _, item := range res.WorldItems {
		if item.DmMembers != nil {
			for _, member := range item.DmMembers.Members {
				userIdMap[member.Id] = struct{}{}
			}
		}
	}
	userIds := make([]string, len(userIdMap))
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
		roomType := database.RoomTypeGroupDM
		isDm := false
		if item.DmMembers != nil {
			roomType = database.RoomTypeDM
			gcMembers = item.DmMembers.Members
			isDm = true
			for _, member := range item.DmMembers.Members {
				if member.Id != string(c.userLogin.ID) {
					name = c.users[member.Id].Name
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
				Name:    &name,
				Members: c.gcMembersToMatrix(isDm, gcMembers),
				Type:    &roomType,
			},
		})

		c.backfillPortal(ctx, item)
	}
}
