package connector

import (
	"context"

	"maunium.net/go/mautrix/bridgev2"
)

func (c *GChatClient) HandleMatrixMessage(ctx context.Context, msg *bridgev2.MatrixMessage) (message *bridgev2.MatrixMessageResponse, err error) {
	return nil, nil
}
