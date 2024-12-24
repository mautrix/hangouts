package msgconv

import (
	"context"

	"maunium.net/go/mautrix/event"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
	"go.mau.fi/mautrix-googlechat/pkg/msgconv/matrixfmt"
)

func (mc *MessageConverter) ToGChat(
	ctx context.Context,
	content *event.MessageEventContent,
) (string, []*proto.Annotation) {
	body, annotations := matrixfmt.Parse(ctx, mc.matrixFmtParams, content)
	return body, annotations
}
