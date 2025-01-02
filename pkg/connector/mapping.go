package connector

import (
	"context"
	"io"
	"net/http"

	"go.mau.fi/util/ptr"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func (c *GChatClient) makeAvatar(avatarURL string) *bridgev2.Avatar {
	return &bridgev2.Avatar{
		ID: networkid.AvatarID(avatarURL),
		Get: func(ctx context.Context) ([]byte, error) {
			resp, err := http.Get(avatarURL)
			if err != nil {
				return nil, err
			}
			data, err := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return data, err
		},
		Remove: avatarURL == "",
	}
}

func (c *GChatClient) makeUserInfo(user *proto.User) *bridgev2.UserInfo {
	if user == nil {
		return nil
	}
	return &bridgev2.UserInfo{
		Name:   &user.Name,
		Avatar: c.makeAvatar(user.AvatarUrl),
	}
}

func (c *GChatClient) gcMembersToMatrix(isDm bool, gcMembers []*proto.UserId) *bridgev2.ChatMemberList {
	var otherUserId string
	memberMap := map[networkid.UserID]bridgev2.ChatMember{}
	for _, gcMember := range gcMembers {
		userId := networkid.UserID(gcMember.Id)
		if isDm && gcMember.Id != string(c.userLogin.ID) {
			otherUserId = gcMember.Id
		}
		isMe := gcMember.Id == string(c.userLogin.ID)
		member := bridgev2.ChatMember{
			EventSender: bridgev2.EventSender{
				IsFromMe: isMe,
				Sender:   userId,
			},
		}
		if isMe {
			member.PowerLevel = ptr.Ptr(50)
		}
		user := c.users[gcMember.Id]
		if user != nil {
			member.UserInfo = c.makeUserInfo(user)
		}
		memberMap[userId] = member
	}

	return &bridgev2.ChatMemberList{
		IsFull:      true,
		MemberMap:   memberMap,
		OtherUserID: networkid.UserID(otherUserId),
	}
}
