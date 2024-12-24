package msgconv

import (
	"context"

	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/id"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow"
	"go.mau.fi/mautrix-googlechat/pkg/msgconv/matrixfmt"
)

type contextKey int

const (
	contextKeyPortal contextKey = iota
	contextKeyClient
	contextKeyIntent
)

type MessageConverter struct {
	client *gchatmeow.Client

	matrixFmtParams *matrixfmt.HTMLParser
}

func NewMessageConverter(br *bridgev2.Bridge, client *gchatmeow.Client) *MessageConverter {
	return &MessageConverter{
		client: client,

		matrixFmtParams: &matrixfmt.HTMLParser{
			GetUIDFromMXID: func(ctx context.Context, userID id.UserID) string {
				parsed, ok := br.Matrix.ParseGhostMXID(userID)
				if ok {
					return string(parsed)
				}
				user, _ := br.GetExistingUserByMXID(ctx, userID)
				if user != nil {
					preferredLogin, _, _ := getPortal(ctx).FindPreferredLogin(ctx, user, true)
					if preferredLogin != nil {
						return string(preferredLogin.ID)
					}
				}
				return ""
			},
		},
	}
}

func getPortal(ctx context.Context) *bridgev2.Portal {
	return ctx.Value(contextKeyPortal).(*bridgev2.Portal)
}
