package connector

import (
	"context"
	"time"

	"google.golang.org/protobuf/encoding/prototext"
	"maunium.net/go/mautrix/bridgev2"
	"maunium.net/go/mautrix/bridgev2/database"
	"maunium.net/go/mautrix/bridgev2/networkid"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto"
)

var (
	_ bridgev2.ReactionHandlingNetworkAPI = (*GChatClient)(nil)
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

	var plainGroupId string
	if groupId.GetDmId() != nil {
		plainGroupId = groupId.GetDmId().DmId
	} else {
		plainGroupId = groupId.GetSpaceId().SpaceId
	}

	var annotations []*proto.Annotation
	messageInfo := &proto.MessageInfo{
		AcceptFormatAnnotations: true,
	}

	if msg.Content.MsgType.IsMedia() {
		data, err := c.userLogin.Bridge.Bot.DownloadMedia(ctx, msg.Content.URL, msg.Content.File)
		if err != nil {
			return nil, err
		}
		metadata, err := c.client.UploadFile(ctx, data, plainGroupId, msg.Content.FileName, msg.Content.Info.MimeType)
		if err != nil {
			return nil, err
		}
		annotations = []*proto.Annotation{
			{
				Type:           proto.AnnotationType_UPLOAD_METADATA,
				ChipRenderType: proto.Annotation_RENDER,
				Metadata: &proto.Annotation_UploadMetadata{
					UploadMetadata: metadata,
				},
			},
		}
	}

	if msg.ReplyTo != nil {
		replyToId := string(msg.ReplyTo.ID)
		topicId := replyToId
		if msg.ThreadRoot != nil {
			topicId = string(msg.ThreadRoot.ID)
		}
		messageInfo.ReplyTo = &proto.SendReplyTarget{
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
			CreateTime: msg.ReplyTo.Timestamp.UnixMicro(),
		}
	}

	var msgID string
	var timestamp int64

	textBody := msg.Content.Body
	text, entities := c.msgConv.ToGChat(ctx, msg.Content)

	if entities != nil {
		textBody = text
		annotations = entities
	}

	if msg.ThreadRoot != nil {
		threadId := string(msg.ThreadRoot.ID)
		req := &proto.CreateMessageRequest{
			ParentId: &proto.MessageParentId{
				Parent: &proto.MessageParentId_TopicId{
					TopicId: &proto.TopicId{
						GroupId: groupId,
						TopicId: threadId,
					},
				},
			},
			LocalId:     string(msg.Event.ID),
			TextBody:    textBody,
			Annotations: annotations,
			MessageInfo: messageInfo,
		}
		res, err := c.client.CreateMessage(ctx, req)
		if err != nil {
			return nil, err
		}
		msgID = res.Message.Id.MessageId
		timestamp = res.Message.CreateTime
	} else {
		req := &proto.CreateTopicRequest{
			GroupId:     groupId,
			TextBody:    textBody,
			Annotations: annotations,
			MessageInfo: messageInfo,
		}
		res, err := c.client.CreateTopic(ctx, req)
		if err != nil {
			return nil, err
		}
		msgID = res.Topic.Id.TopicId
		timestamp = res.Topic.CreateTimeUsec
	}

	msg.AddPendingToIgnore(networkid.TransactionID(msgID))
	return &bridgev2.MatrixMessageResponse{
		DB: &database.Message{
			ID:        networkid.MessageID(msgID),
			Timestamp: time.UnixMicro(timestamp),
		},
		RemovePending: networkid.TransactionID(msgID),
	}, nil
}

func (c *GChatClient) PreHandleMatrixReaction(_ context.Context, msg *bridgev2.MatrixReaction) (bridgev2.MatrixReactionPreResponse, error) {
	emoji := msg.Content.RelatesTo.Key
	return bridgev2.MatrixReactionPreResponse{
		SenderID: networkid.UserID(c.userLogin.ID),
		EmojiID:  networkid.EmojiID(emoji),
		Emoji:    emoji,
	}, nil
}

func (c *GChatClient) HandleMatrixReaction(ctx context.Context, msg *bridgev2.MatrixReaction) (*database.Reaction, error) {
	return nil, c.doHandleMatrixReaction(ctx, msg.Portal,
		string(msg.TargetMessage.ThreadRoot),
		string(msg.TargetMessage.ID), msg.PreHandleResp.Emoji, proto.UpdateReactionRequest_ADD)
}

func (c *GChatClient) HandleMatrixReactionRemove(ctx context.Context, msg *bridgev2.MatrixReactionRemove) error {
	dbMsg, err := c.userLogin.Bridge.DB.Message.GetLastPartByID(ctx, c.userLogin.ID, msg.TargetReaction.MessageID)
	if err != nil {
		return err
	}
	var topicId string
	if dbMsg != nil {
		topicId = string(dbMsg.ThreadRoot)
	}
	return c.doHandleMatrixReaction(ctx, msg.Portal,
		topicId,
		string(msg.TargetReaction.MessageID), string(msg.TargetReaction.EmojiID), proto.UpdateReactionRequest_REMOVE)
}

func (c *GChatClient) doHandleMatrixReaction(ctx context.Context, portal *bridgev2.Portal, topicId, messageId, emoji string, typ proto.UpdateReactionRequest_ReactionUpdateType) error {
	groupId, err := portalToGroupId(portal)
	if err != nil {
		return err
	}

	if topicId == "" {
		topicId = messageId
	}
	_, err = c.client.UpdateReaction(ctx, &proto.UpdateReactionRequest{
		MessageId: &proto.MessageId{
			ParentId: &proto.MessageParentId{
				Parent: &proto.MessageParentId_TopicId{
					TopicId: &proto.TopicId{
						GroupId: groupId,
						TopicId: topicId,
					},
				},
			},
			MessageId: messageId,
		},
		Emoji: &proto.Emoji{
			Content: &proto.Emoji_Unicode{
				Unicode: emoji,
			},
		},
		Type: typ,
	})
	return err
}
