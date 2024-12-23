package gchatfmt_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
	"go.mau.fi/mautrix-googlechat/pkg/msgconv/gchatfmt"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		ins  string
		ine  []*proto.Annotation
		body string
		html string
	}{
		{
			name: "plain",
			ins:  "Hello world!",
			body: "Hello world!",
		},
		{
			name: "bold italic strike underline",
			ins:  "a b i s u z",
			ine: []*proto.Annotation{
				gchatfmt.MakeAnnotation(2, 1, proto.FormatMetadata_BOLD),
				gchatfmt.MakeAnnotation(4, 1, proto.FormatMetadata_ITALIC),
				gchatfmt.MakeAnnotation(6, 1, proto.FormatMetadata_STRIKE),
				gchatfmt.MakeAnnotation(8, 1, proto.FormatMetadata_UNDERLINE),
			},
			body: "a b i s u z",
			html: "a <strong>b</strong> <em>i</em> <del>s</del> <u>u</u> z",
		},
		{
			name: "emoji",
			ins:  "🎆 a b z",
			ine: []*proto.Annotation{
				gchatfmt.MakeAnnotation(5, 1, proto.FormatMetadata_BOLD),
			},
			body: "🎆 a b z",
			html: "🎆 a <strong>b</strong> z",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			msg := &proto.Message{
				TextBody:    test.ins,
				Annotations: test.ine,
			}
			parsed := gchatfmt.Parse(context.TODO(), nil, msg)
			assert.Equal(t, test.body, parsed.Body)
			assert.Equal(t, test.html, parsed.FormattedBody)
		})
	}
}
