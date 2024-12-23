package matrixfmt_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"maunium.net/go/mautrix/event"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
	"go.mau.fi/mautrix-googlechat/pkg/msgconv/gchatfmt"
	"go.mau.fi/mautrix-googlechat/pkg/msgconv/matrixfmt"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		out  string
		ent  []*proto.Annotation
	}{
		{name: "Plain", in: "Hello, World!", out: "Hello, World!"},
		{name: "Bold", in: "a <strong>b</strong> c", out: "a b c",
			ent: []*proto.Annotation{
				gchatfmt.MakeAnnotation(2, 1, proto.FormatMetadata_BOLD),
			},
		},
	}

	parser := &matrixfmt.HTMLParser{}
	matrixfmt.DebugLog = func(format string, args ...any) {
		fmt.Printf(format, args...)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parsed, entities := matrixfmt.Parse(context.TODO(), parser, &event.MessageEventContent{
				Format:        event.FormatHTML,
				FormattedBody: test.in,
			})
			assert.Equal(t, test.out, parsed)
			assert.Equal(t, test.ent, entities)
		})
	}
}
