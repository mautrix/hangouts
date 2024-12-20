package msgconv

import (
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow"
)

type MessageConverter struct {
	client *gchatmeow.Client
}

func NewMessageConverter(client *gchatmeow.Client) *MessageConverter {
	return &MessageConverter{
		client: client,
	}
}
