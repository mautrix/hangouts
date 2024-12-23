package matrixfmt

import (
	"context"

	"maunium.net/go/mautrix/event"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func Parse(ctx context.Context, parser *HTMLParser, content *event.MessageEventContent) (string, []*proto.Annotation) {
	if content.Format != event.FormatHTML {
		return content.Body, nil
	}
	parseCtx := NewContext(ctx)
	parseCtx.AllowedMentions = content.Mentions
	parsed := parser.Parse(content.FormattedBody, parseCtx)
	if parsed == nil {
		return "", nil
	}
	var bodyRanges []*proto.Annotation
	if len(parsed.Entities) > 0 {
		bodyRanges = make([]*proto.Annotation, len(parsed.Entities))
		for i, ent := range parsed.Entities {
			bodyRanges[i] = ent.Proto()
		}
	}
	return parsed.String.String(), bodyRanges
}
