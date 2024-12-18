package connector

import (
	"maunium.net/go/mautrix/bridgev2/networkid"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func (c *GChatClient) makePortalKey(evt *proto.Event) networkid.PortalKey {
	return networkid.PortalKey{
		ID:       networkid.PortalID(evt.GroupId.String()),
		Receiver: c.userLogin.ID,
	}
}
