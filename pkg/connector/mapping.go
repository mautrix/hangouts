package connector

import (
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/networkid"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func (c *GChatClient) gcMembersToMatrix(isDm bool, gcMembers []*proto.UserId) *bridgev2.ChatMemberList {
	var otherUserId string
	memberMap := map[networkid.UserID]bridgev2.ChatMember{}
	for _, gcMember := range gcMembers {
		userId := networkid.UserID(gcMember.Id)
		if isDm && gcMember.Id != string(c.userLogin.ID) {
			otherUserId = gcMember.Id

		}
		memberMap[userId] = bridgev2.ChatMember{
			EventSender: bridgev2.EventSender{
				IsFromMe: gcMember.Id == string(c.userLogin.ID),
				Sender:   userId,
			},
		}
	}

	return &bridgev2.ChatMemberList{
		MemberMap:   memberMap,
		OtherUserID: networkid.UserID(otherUserId),
	}
}
