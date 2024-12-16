package connector

import (
	"context"
	"strconv"
	"time"

	"google.golang.org/protobuf/encoding/prototext"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"

	"go.mau.fi/util/ptr"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

func portalToGroupId(portal *bridgev2.Portal) (*proto.GroupId, error) {
	groupId := &proto.GroupId{}
	err := prototext.Unmarshal([]byte(portal.ID), groupId)
	if err != nil {
		return nil, err
	}

	return groupId, nil
}

func (c *GChatClient) HandleMatrixMessage(ctx context.Context, msg *bridgev2.MatrixMessage) (message *bridgev2.MatrixMessageResponse, err error) {
	groupId, err := portalToGroupId(msg.Portal)
	if err != nil {
		return nil, err
	}

	var plainGroupId *string
	if groupId.GetDmId() != nil {
		plainGroupId = groupId.GetDmId().DmId
	} else {
		plainGroupId = groupId.GetSpaceId().SpaceId
	}

	var annotations []*proto.Annotation
	var messageInfo *proto.MessageInfo

	if msg.Content.MsgType.IsMedia() {
		data, err := c.userLogin.Bridge.Bot.DownloadMedia(ctx, msg.Content.URL, msg.Content.File)
		if err != nil {
			return nil, err
		}
		metadata, err := c.client.UploadFile(ctx, data, *plainGroupId, msg.Content.FileName, msg.Content.Info.MimeType)
		if err != nil {
			return nil, err
		}
		annotations = []*proto.Annotation{
			{
				Type:           ptr.Ptr(proto.AnnotationType_UPLOAD_METADATA),
				ChipRenderType: ptr.Ptr(proto.Annotation_RENDER),
				Metadata: &proto.Annotation_UploadMetadata{
					UploadMetadata: metadata,
				},
			},
		}
	}

	if msg.ReplyTo != nil {
		replyToId := ptr.Ptr(string(msg.ReplyTo.ID))
		topicId := replyToId
		if msg.ThreadRoot != nil {
			topicId = ptr.Ptr(string(msg.ThreadRoot.ID))
		}
		messageInfo = &proto.MessageInfo{
			AcceptFormatAnnotations: ptr.Ptr(true),
			ReplyTo: &proto.SendReplyTarget{
				Id: &proto.MessageId{
					ParentId: &proto.MessageParentId{
						Parent: &proto.MessageParentId_TopicId{
							TopicId: &proto.TopicId{
								GroupId: groupId,
								TopicId: topicId,
							},
						},
					},
					MessageId: replyToId,
				},
				CreateTime: ptr.Ptr(msg.ReplyTo.Timestamp.UnixMicro()),
			},
		}
	}

	var msgID string

	if msg.ThreadRoot != nil {
		threadId := ptr.Ptr(string(msg.ThreadRoot.ID))
		if messageInfo == nil {
			messageInfo = &proto.MessageInfo{
				AcceptFormatAnnotations: ptr.Ptr(true),
			}
		}
		req := &proto.CreateMessageRequest{
			ParentId: &proto.MessageParentId{
				Parent: &proto.MessageParentId_TopicId{
					TopicId: &proto.TopicId{
						GroupId: groupId,
						TopicId: threadId,
					},
				},
			},
			LocalId:     ptr.Ptr(strconv.FormatInt(time.Now().Unix(), 10)),
			TextBody:    &msg.Content.Body,
			Annotations: annotations,
			MessageInfo: messageInfo,
		}
		res, err := c.client.CreateMessage(ctx, req)
		if err != nil {
			return nil, err
		}
		msgID = *res.Message.LocalId
	} else {
		req := &proto.CreateTopicRequest{
			GroupId:     groupId,
			TextBody:    &msg.Content.Body,
			Annotations: annotations,
			MessageInfo: messageInfo,
		}
		res, err := c.client.CreateTopic(ctx, req)
		if err != nil {
			return nil, err
		}
		msgID = *res.Topic.Id.TopicId
	}

	msg.AddPendingToIgnore(networkid.TransactionID(msgID))
	return &bridgev2.MatrixMessageResponse{
		DB: &database.Message{
			ID: networkid.MessageID(msgID),
		},
		RemovePending: networkid.TransactionID(msgID),
	}, nil
}
