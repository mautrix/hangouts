package pblite_test

import (
	"log"
	"os"
	"testing"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/pblite"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatproto"
)

func TestPblite(t *testing.T) {
	data, err := os.ReadFile("test_data.json")
	if err != nil {
		log.Fatal(err)
	}

	schema := &gchatproto.WebChannelEvent{}
	err = pblite.Unmarshal(data, schema)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(schema.Metadata)
}
