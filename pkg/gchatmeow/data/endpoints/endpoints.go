package endpoints

type APIEndpoint string

const (
	WEBCHANNEL_BASE_URL = "https://chat.google.com/u/0/webchannel"
	WEBCHANNEL_REGISTER = WEBCHANNEL_BASE_URL + "/register"
	WEBCHANNEL_EVENTS   = WEBCHANNEL_BASE_URL + "/events"

	CHAT_BASE_URL                  = "https://chat.google.com/"
	API_BASE_URL                   = "https://chat.google.com/u/0/api"
	PAGINATED_WORLD    APIEndpoint = "/paginated_world"
	LIST_MEMBERS       APIEndpoint = "/list_members"
	CREATE_TOPIC       APIEndpoint = "/create_topic"
	LIST_TOPICS        APIEndpoint = "/list_topics"
	UPDATE_REACTION    APIEndpoint = "/update_reaction"
	DELETE_MESSAGE     APIEndpoint = "/delete_message"
	EDIT_MESSAGE       APIEndpoint = "/edit_message"
	CREATE_GROUP       APIEndpoint = "/create_group"
	CREATE_MEMBERSHIP  APIEndpoint = "/create_membership"
	REMOVE_MEMBERSHIPS APIEndpoint = "/remove_memberships"
	GET_GROUP          APIEndpoint = "/get_group"
	MARK_AS_UNREAD     APIEndpoint = "/set_mark_as_unread_timestamp"
	CREATE_DM_EXTENDED APIEndpoint = "/create_dm_extended"

	UPLOADS = CHAT_BASE_URL + "uploads"

	MOLE_BASE_URL = "https://chat.google.com/u/0/mole/world"
)
