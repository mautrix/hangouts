package gchatfmt_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mau.fi/util/ptr"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
	"go.mau.fi/mautrix-googlechat/pkg/msgconv/gchatfmt"
)

func makeAnnotation(start, length int32, format proto.FormatMetadata_FormatType) *proto.Annotation {
	return &proto.Annotation{
		Type:           proto.AnnotationType_FORMAT_DATA.Enum(),
		StartIndex:     ptr.Ptr(start),
		Length:         ptr.Ptr(length),
		ChipRenderType: ptr.Ptr(proto.Annotation_DO_NOT_RENDER),
		Metadata: &proto.Annotation_FormatMetadata{
			FormatMetadata: &proto.FormatMetadata{
				FormatType: format.Enum(),
			},
		},
	}
}

func TestParse(t *testing.T) {
	assert.Equal(t, 1, 1)

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
				makeAnnotation(2, 1, proto.FormatMetadata_BOLD),
				makeAnnotation(4, 1, proto.FormatMetadata_ITALIC),
				makeAnnotation(6, 1, proto.FormatMetadata_STRIKE),
				makeAnnotation(8, 1, proto.FormatMetadata_UNDERLINE),
			},
			body: "a b i s u z",
			html: "a <strong>b</strong> <em>i</em> <del>s</del> <u>u</u> z",
		},
		{
			name: "emoji",
			ins:  "🎆 a b z",
			ine: []*proto.Annotation{
				makeAnnotation(5, 1, proto.FormatMetadata_BOLD),
			},
			body: "🎆 a b z",
			html: "🎆 a <strong>b</strong> z",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			msg := &proto.Message{
				TextBody:    ptr.Ptr(test.ins),
				Annotations: test.ine,
			}
			parsed := gchatfmt.Parse(context.TODO(), nil, msg)
			assert.Equal(t, test.body, parsed.Body)
			assert.Equal(t, test.html, parsed.FormattedBody)
		})
	}
}
