package connector

import (
	"context"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/bridgev2/networkid"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

type PortalMetadata struct {
	Revision int64
}

func (c *GChatClient) setPortalRevision(ctx context.Context, evt *proto.Event) {
	portalKey := networkid.PortalKey{
		ID:       networkid.PortalID(evt.GroupId.String()),
		Receiver: c.userLogin.ID,
	}
	portal, err := c.userLogin.Bridge.GetPortalByKey(ctx, portalKey)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("Failed to get portal by key")
		return
	}
	metadata := portal.Metadata.(*PortalMetadata)
	revision := evt.GetGroupRevision()
	if revision != nil {
		metadata.Revision = revision.Timestamp
		err = portal.Save(ctx)
		zerolog.Ctx(ctx).Err(err).Msg("Failed to update portal revision in database")
	}
}

func (c *GChatClient) backfillPortal(ctx context.Context, item *proto.WorldItemLite) {
	portalKey := networkid.PortalKey{
		ID:       networkid.PortalID(item.GroupId.String()),
		Receiver: c.userLogin.ID,
	}
	portal, err := c.userLogin.Bridge.GetPortalByKey(ctx, portalKey)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Msg("Failed to get portal by key")
		return
	}
	metadata := portal.Metadata.(*PortalMetadata)

	if metadata.Revision < item.GroupRevision.Timestamp {
		res, err := c.client.CatchUpGroup(ctx, &proto.CatchUpGroupRequest{
			GroupId: item.GroupId,
			Range: &proto.CatchUpRange{
				FromRevisionTimestamp: metadata.Revision,
				ToRevisionTimestamp:   item.GroupRevision.Timestamp,
			},
			PageSize:   int32(c.userLogin.Bridge.Config.Backfill.Queue.BatchSize),
			CutoffSize: int32(c.userLogin.Bridge.Config.Backfill.MaxCatchupMessages),
		})
		if err != nil {
			zerolog.Ctx(ctx).Err(err).Msg("Failed to catch up portal " + item.GroupId.String())
			return
		}

		for _, event := range res.Events {
			for _, evt := range c.client.SplitEventBodies(event) {
				c.onStreamEvent(ctx, evt)
			}
		}
	}
}
