//lint:file-ignore U1000 ignore

package gchatmeow_test

import (
	"log"
	"os"
	"testing"

	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/cookies"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/data/query"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/debug"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/event"
	"go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatproto"

	"go.mau.fi/util/random"
)

var cli *gchatmeow.Client

func TestXClientLogin(t *testing.T) {
	cookieStr, err := os.ReadFile("cookies.txt")
	if err != nil {
		log.Fatal(err)
	}
	cookieStruct := cookies.NewCookiesFromString(string(cookieStr))

	clientOptions := gchatmeow.ClientOpts{
		Cookies:      cookieStruct,
		EventHandler: eventHandler,
	}
	cli = gchatmeow.NewClient(&clientOptions, debug.NewLogger())
	cli.SetEventHandler(eventHandler)

	initialData, err := cli.LoadMessagesPage()
	if err != nil {
		log.Fatal(err)
	}

	groupListData := initialData.GroupList
	currentUser := initialData.CurrentUser.Data
	cli.Logger.Info().
		Str("full_name", currentUser.Fullname).
		Str("user_id", currentUser.IdDeprecated).
		Str("email", currentUser.Email).
		Int("spaces", len(groupListData.Spaces)).
		Int("single_dms", len(groupListData.SingleDms)).
		Msg("Successfully authenticated")

	err = cli.Connect()
	if err != nil {
		log.Fatalf("failed to connect with realtime client (%s)", err.Error())
	}

	wait := make(chan struct{})
	<-wait
}

func testUploadMedia() (string, *gchatproto.UploadMetadata) {
	mediaBytes, err := os.ReadFile("test_data/testimage1.jpg")
	if err != nil {
		log.Fatal(err)
	}

	messageId := random.String(11)
	uploadQuery := query.UploadMediaQuery{
		GroupID:         "my-group-id",
		TopicID:         messageId,
		MessageID:       messageId,
		Otr:             "false",
		TranscodedVideo: "false",
		UploadType:      "ATTACHMENT",
	}
	mediaUpload, err := cli.UploadMedia("testimage1.jpg", "image/jpeg", mediaBytes, uploadQuery)
	if err != nil {
		log.Fatal(err)
	}

	cli.Logger.Info().Any("media_metadata", mediaUpload).Msg("Successfully uploaded media")
	return messageId, mediaUpload
}

func testUploadVideo() (string, *gchatproto.UploadMetadata) {
	mediaBytes, err := os.ReadFile("test_data/testvideo1.mp4")
	if err != nil {
		log.Fatal(err)
	}

	messageId := random.String(11)
	uploadQuery := query.UploadMediaQuery{
		GroupID:         "my-group-id",
		TopicID:         messageId,
		MessageID:       messageId,
		Otr:             "false",
		TranscodedVideo: "false",
		UploadType:      "ATTACHMENT",
	}
	mediaUpload, err := cli.UploadMedia("testvideo1.mp4", "video/mp4", mediaBytes, uploadQuery)
	if err != nil {
		log.Fatal(err)
	}

	cli.Logger.Info().Any("media_metadata", mediaUpload).Msg("Successfully uploaded video")
	return messageId, mediaUpload
}

func testSendImage() {
	messageId, mediaData := testUploadMedia()
	payload := &gchatproto.CreateTopicRequest{
		GroupId: &gchatproto.GroupId{
			Id: &gchatproto.GroupId_SpaceId{
				SpaceId: &gchatproto.SpaceId{
					SpaceId: "my-space-id",
				},
			},
		},
		Annotations: []*gchatproto.Annotation{
			{
				Type:       gchatproto.AnnotationType_UPLOAD_METADATA,
				StartIndex: 0,
				Length:     0,
				Metadata: &gchatproto.Annotation_UploadMetadata{
					UploadMetadata: mediaData,
				},
			},
		},
		TextBody:          "testing sending image",
		HistoryV2:         true,
		TopicAndMessageId: messageId,
	}

	resp, err := cli.SendMessage(payload)
	if err != nil {
		log.Fatal(err)
	}

	cli.Logger.Info().Any("resp", resp).Msg("Sent test image!")
	os.Exit(1)
}

func testSendVideo() {
	messageId, mediaData := testUploadVideo()
	payload := &gchatproto.CreateTopicRequest{
		GroupId: &gchatproto.GroupId{
			Id: &gchatproto.GroupId_SpaceId{
				SpaceId: &gchatproto.SpaceId{
					SpaceId: "my-space-id",
				},
			},
		},
		Annotations: []*gchatproto.Annotation{
			{
				Type:       gchatproto.AnnotationType_UPLOAD_METADATA,
				StartIndex: 0,
				Length:     0,
				Metadata: &gchatproto.Annotation_UploadMetadata{
					UploadMetadata: mediaData,
				},
			},
		},
		TextBody:          "testing sending video",
		HistoryV2:         true,
		TopicAndMessageId: messageId,
	}

	resp, err := cli.SendMessage(payload)
	if err != nil {
		log.Fatal(err)
	}

	cli.Logger.Info().Any("resp", resp).Msg("Sent test video!")
	os.Exit(1)
}

func testListMessages() {
	payload := &gchatproto.ListTopicsRequest{
		PageSizeForTopics: 20,
		GroupId: &gchatproto.GroupId{
			Id: &gchatproto.GroupId_SpaceId{
				SpaceId: &gchatproto.SpaceId{
					SpaceId: "my-space-id",
				},
			},
		},
	}

	response, err := cli.ListMessages(payload)
	if err != nil {
		log.Fatal(err)
	}

	for _, msg := range response.Topics {
		for _, reply := range msg.Replies { // ????
			cli.Logger.Info().
				Str("id", msg.Id.GetTopicId()).
				Str("text", reply.TextBody).
				Msg("Found message")
		}
	}
	os.Exit(1)
}

func testGetPaginatedWorlds() {
	response, err := cli.GetPaginatedWorlds(nil)
	if err != nil {
		log.Fatal(err)
	}

	worlds := response.WorldItems
	for _, world := range worlds {
		cli.Logger.Info().Any("GroupNameInfo", world.GroupNameInfo).Msg("World name")
	}

	os.Exit(1)
}

func testSendMessage() {
	payload := &gchatproto.CreateTopicRequest{
		GroupId: &gchatproto.GroupId{
			Id: &gchatproto.GroupId_SpaceId{
				SpaceId: &gchatproto.SpaceId{
					SpaceId: "my-space-id",
				},
			},
		},
		TextBody:  "testing this",
		HistoryV2: true,
	}

	response, err := cli.SendMessage(payload)
	if err != nil {
		log.Fatal(err)
	}

	cli.Logger.Info().Any("message_response", response).Msg("Sent test message")
	os.Exit(1)
}

func testListMembers() {
	payload := &gchatproto.ListMembersRequest{
		GroupId: &gchatproto.GroupId{
			Id: &gchatproto.GroupId_SpaceId{
				SpaceId: &gchatproto.SpaceId{
					SpaceId: "the-space-id",
				},
			},
		},
	}

	response, err := cli.ListMembers(payload)
	if err != nil {
		log.Fatal(err)
	}

	for _, memberShip := range response.Memberships {
		cli.Logger.Info().Any("member", memberShip).Msg("List Members Response")
	}
	os.Exit(1)
}

func eventHandler(data any) {
	switch evData := data.(type) {
	case *event.MessageEvent:
		cli.Logger.Info().Any("event_data", evData).Msg("Received message event")
	case *event.SessionReadyEvent: // this event is triggered everytime the realtime client opens a connection, including reconnects.
		cli.Logger.Info().Msg("Realtime session is ready")
	default:
		cli.Logger.Info().Any("event_data", evData).Msg("Received unhandled event")
	}
}
