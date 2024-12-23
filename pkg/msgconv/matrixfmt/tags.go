package matrixfmt

import (
	"fmt"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

type BodyRangeValue interface {
	String() string
	Format(message string) string
	Proto() proto.FormatMetadata_FormatType
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

func (s Style) Proto() proto.FormatMetadata_FormatType {
	return proto.FormatMetadata_FormatType(s)
}

func (s Style) String() string {
	return fmt.Sprintf("Style(%d)", s)
}

func (s Style) Format(message string) string {
	return message
}
