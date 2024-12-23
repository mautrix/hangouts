package gchatfmt

import "go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"

func MakeAnnotation(start, length int32, format proto.FormatMetadata_FormatType) *proto.Annotation {
	return &proto.Annotation{
		Type:           proto.AnnotationType_FORMAT_DATA,
		StartIndex:     start,
		Length:         length,
		ChipRenderType: proto.Annotation_DO_NOT_RENDER,
		Metadata: &proto.Annotation_FormatMetadata{
			FormatMetadata: &proto.FormatMetadata{
				FormatType: format,
			},
		},
	}

}
