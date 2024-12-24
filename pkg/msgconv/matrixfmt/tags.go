package matrixfmt

import (
	"fmt"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

type BodyRangeValue interface {
	String() string
	Format(message string) string
	Proto() proto.MetadataAssociatedValue
}

type Mention struct {
	ID string
}

func (m Mention) String() string {
	return fmt.Sprintf("Mention{ID: (%s)}", m.ID)
}

func (m Mention) Proto() proto.MetadataAssociatedValue {
	return &proto.Annotation_UserMentionMetadata{
		UserMentionMetadata: &proto.UserMentionMetadata{
			Type: proto.UserMentionMetadata_MENTION,
			Id: &proto.UserId{
				Id: m.ID,
			},
		},
	}
}

func (m Mention) Format(message string) string {
	return message
}

type Style int

const (
	StyleNone Style = iota
	StyleBold
	StyleItalic
	StyleStrikethrough
	StyleSourceCode
	StyleMonospace // 5
	StyleHidden
	StyleMonospaceBlock
	StyleUnderline
	StyleFontColor
)

func (s Style) Proto() proto.MetadataAssociatedValue {
	return &proto.Annotation_FormatMetadata{
		FormatMetadata: &proto.FormatMetadata{
			FormatType: proto.FormatMetadata_FormatType(s),
		},
	}
}

func (s Style) String() string {
	return fmt.Sprintf("Style(%d)", s)
}

func (s Style) Format(message string) string {
	return message
}
