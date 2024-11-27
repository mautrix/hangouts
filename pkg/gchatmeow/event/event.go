package event

import "go.mau.fi/mautrix-googlechat/pkg/gchatmeow/proto/gchatproto"

/*
	Refactored event structs for aesthetics ;)
*/

type MessageEvent struct {
	*gchatproto.EventBody_MessagePosted
}

type GroupViewed struct {
	*gchatproto.EventBody_GroupViewed
}

type TopicViewed struct {
	*gchatproto.EventBody_TopicViewed
}

type GroupUpdated struct {
	*gchatproto.EventBody_GroupUpdated
}

type TopicMuteChanged struct {
	*gchatproto.EventBody_TopicMuteChanged
}

type UserSettingsChanged struct {
	*gchatproto.EventBody_UserSettingsChanged
}

type GroupStarred struct {
	*gchatproto.EventBody_GroupStarred
}

type WebPushNotification struct {
	*gchatproto.EventBody_WebPushNotification
}

type GroupUnreadSubscribedTopicCountUpdatedEvent struct {
	*gchatproto.EventBody_GroupUnreadSubscribedTopicCountUpdatedEvent
}

type InviteCountUpdated struct {
	*gchatproto.EventBody_InviteCountUpdated
}

type MembershipChanged struct {
	*gchatproto.EventBody_MembershipChanged
}

type GroupHideChanged struct {
	*gchatproto.EventBody_GroupHideChanged
}

type DriveAclFixProcessed struct {
	*gchatproto.EventBody_DriveAclFixProcessed
}

type GroupNotificationSettingsUpdated struct {
	*gchatproto.EventBody_GroupNotificationSettingsUpdated
}

type MessageDeleted struct {
	*gchatproto.EventBody_MessageDeleted
}

type RetentionSettingsUpdated struct {
	*gchatproto.EventBody_RetentionSettingsUpdated
}

type TopicCreated struct {
	*gchatproto.EventBody_TopicCreated
}

type MessageReaction struct {
	*gchatproto.EventBody_MessageReaction
}

type UserStatusUpdatedEvent struct {
	*gchatproto.EventBody_UserStatusUpdatedEvent
}

type WorkingHoursSettingsUpdatedEvent struct {
	*gchatproto.EventBody_WorkingHoursSettingsUpdatedEvent
}

type MessageSmartRepliesEvent struct {
	*gchatproto.EventBody_MessageSmartRepliesEvent
}

type TypingStateChangedEvent struct {
	*gchatproto.EventBody_TypingStateChangedEvent
}

type GroupDeletedEvent struct {
	*gchatproto.EventBody_GroupDeletedEvent
}

type BlockStateChangedEvent struct {
	*gchatproto.EventBody_BlockStateChangedEvent
}

type ClearHistoryEvent struct {
	*gchatproto.EventBody_ClearHistoryEvent
}

type GroupSortTimestampChangedEvent struct {
	*gchatproto.EventBody_GroupSortTimestampChangedEvent
}

type MarkAsUnreadEvent struct {
	*gchatproto.EventBody_MarkAsUnreadEvent
}

type ReadReceiptChanged struct {
	*gchatproto.EventBody_ReadReceiptChanged
}

type GroupNoOpEvent struct {
	*gchatproto.EventBody_GroupNoOpEvent
}

type UserNoOpEvent struct {
	*gchatproto.EventBody_UserNoOpEvent
}

type UserDenormalizedGroupUpdatedEvent struct {
	*gchatproto.EventBody_UserDenormalizedGroupUpdatedEvent
}

type NotificationsCardEvent struct {
	*gchatproto.EventBody_NotificationsCardEvent
}

type UserHubAvailabilityEvent struct {
	*gchatproto.EventBody_UserHubAvailabilityEvent
}

type PresenceSharedUpdatedEvent struct {
	*gchatproto.EventBody_PresenceSharedUpdatedEvent
}

type UserOwnershipUpdatedEvent struct {
	*gchatproto.EventBody_UserOwnershipUpdatedEvent
}

type GroupScopedCapabilitiesEvent struct {
	*gchatproto.EventBody_GroupScopedCapabilitiesEvent
}

type MeetEvent struct {
	*gchatproto.EventBody_MeetEvent
}

type GroupUnreadThreadStateUpdatedEvent struct {
	*gchatproto.EventBody_GroupUnreadThreadStateUpdatedEvent
}

type WebchannelCheckEvent struct {
	*gchatproto.EventBody_WebchannelCheckEvent
}

type RecurringDndSettingsUpdatedEvent struct {
	*gchatproto.EventBody_RecurringDndSettingsUpdatedEvent
}

type MessageLabelsUpdatedEvent struct {
	*gchatproto.EventBody_MessageLabelsUpdatedEvent
}

type MessageReactionsSummary struct {
	*gchatproto.EventBody_MessageReactionsSummary
}

type GroupDefaultSortOrderUpdatedEvent struct {
	*gchatproto.EventBody_GroupDefaultSortOrderUpdatedEvent
}

type TopicLabelEvent struct {
	*gchatproto.EventBody_TopicLabelEvent
}

type StringSortOrderUpdatedEvent struct {
	*gchatproto.EventBody_StringSortOrderUpdatedEvent
}

type MessageLabelEvent struct {
	*gchatproto.EventBody_MessageLabelEvent
}

type GroupLabelEvent struct {
	*gchatproto.EventBody_GroupLabelEvent
}

type RosterSectionEvent struct {
	*gchatproto.EventBody_RosterSectionEvent
}

type BadgeCountUpdatedEvent struct {
	*gchatproto.EventBody_BadgeCountUpdatedEvent
}

type WorldRefreshedEvent struct {
	*gchatproto.EventBody_WorldRefreshedEvent
}

type JoinRequestedEvent struct {
	*gchatproto.EventBody_JoinRequestedEvent
}

type GroupEntityEvent struct {
	*gchatproto.EventBody_GroupEntityEvent
}

type MessageDetectedIntentEvent struct {
	*gchatproto.EventBody_MessageDetectedIntentEvent
}

type TopicMetadataUpdatedEvent struct {
	*gchatproto.EventBody_TopicMetadataUpdatedEvent
}

type GroupReadStateUpdatedEvent struct {
	*gchatproto.EventBody_GroupReadStateUpdatedEvent
}

type SessionReadyEvent struct{}

func PrettifyEvent(evBody *gchatproto.EventBody) any {
	var prettifiedEvent any
	switch event := evBody.Type.(type) {
	case *gchatproto.EventBody_MessagePosted:
		prettifiedEvent = &MessageEvent{event}
	case *gchatproto.EventBody_GroupViewed:
		prettifiedEvent = &GroupViewed{event}
	case *gchatproto.EventBody_TopicViewed:
		prettifiedEvent = &TopicViewed{event}
	case *gchatproto.EventBody_GroupUpdated:
		prettifiedEvent = &GroupUpdated{event}
	case *gchatproto.EventBody_TopicMuteChanged:
		prettifiedEvent = &TopicMuteChanged{event}
	case *gchatproto.EventBody_UserSettingsChanged:
		prettifiedEvent = &UserSettingsChanged{event}
	case *gchatproto.EventBody_GroupStarred:
		prettifiedEvent = &GroupStarred{event}
	case *gchatproto.EventBody_WebPushNotification:
		prettifiedEvent = &WebPushNotification{event}
	case *gchatproto.EventBody_GroupUnreadSubscribedTopicCountUpdatedEvent:
		prettifiedEvent = &GroupUnreadSubscribedTopicCountUpdatedEvent{event}
	case *gchatproto.EventBody_InviteCountUpdated:
		prettifiedEvent = &InviteCountUpdated{event}
	case *gchatproto.EventBody_MembershipChanged:
		prettifiedEvent = &MembershipChanged{event}
	case *gchatproto.EventBody_GroupHideChanged:
		prettifiedEvent = &GroupHideChanged{event}
	case *gchatproto.EventBody_DriveAclFixProcessed:
		prettifiedEvent = &DriveAclFixProcessed{event}
	case *gchatproto.EventBody_GroupNotificationSettingsUpdated:
		prettifiedEvent = &GroupNotificationSettingsUpdated{event}
	case *gchatproto.EventBody_MessageDeleted:
		prettifiedEvent = &MessageDeleted{event}
	case *gchatproto.EventBody_RetentionSettingsUpdated:
		prettifiedEvent = &RetentionSettingsUpdated{event}
	case *gchatproto.EventBody_TopicCreated:
		prettifiedEvent = &TopicCreated{event}
	case *gchatproto.EventBody_MessageReaction:
		prettifiedEvent = &MessageReaction{event}
	case *gchatproto.EventBody_UserStatusUpdatedEvent:
		prettifiedEvent = &UserStatusUpdatedEvent{event}
	case *gchatproto.EventBody_WorkingHoursSettingsUpdatedEvent:
		prettifiedEvent = &WorkingHoursSettingsUpdatedEvent{event}
	case *gchatproto.EventBody_MessageSmartRepliesEvent:
		prettifiedEvent = &MessageSmartRepliesEvent{event}
	case *gchatproto.EventBody_TypingStateChangedEvent:
		prettifiedEvent = &TypingStateChangedEvent{event}
	case *gchatproto.EventBody_GroupDeletedEvent:
		prettifiedEvent = &GroupDeletedEvent{event}
	case *gchatproto.EventBody_BlockStateChangedEvent:
		prettifiedEvent = &BlockStateChangedEvent{event}
	case *gchatproto.EventBody_ClearHistoryEvent:
		prettifiedEvent = &ClearHistoryEvent{event}
	case *gchatproto.EventBody_GroupSortTimestampChangedEvent:
		prettifiedEvent = &GroupSortTimestampChangedEvent{event}
	case *gchatproto.EventBody_MarkAsUnreadEvent:
		prettifiedEvent = &MarkAsUnreadEvent{event}
	case *gchatproto.EventBody_ReadReceiptChanged:
		prettifiedEvent = &ReadReceiptChanged{event}
	case *gchatproto.EventBody_GroupNoOpEvent:
		prettifiedEvent = &GroupNoOpEvent{event}
	case *gchatproto.EventBody_UserNoOpEvent:
		prettifiedEvent = &UserNoOpEvent{event}
	case *gchatproto.EventBody_UserDenormalizedGroupUpdatedEvent:
		prettifiedEvent = &UserDenormalizedGroupUpdatedEvent{event}
	case *gchatproto.EventBody_NotificationsCardEvent:
		prettifiedEvent = &NotificationsCardEvent{event}
	case *gchatproto.EventBody_UserHubAvailabilityEvent:
		prettifiedEvent = &UserHubAvailabilityEvent{event}
	case *gchatproto.EventBody_PresenceSharedUpdatedEvent:
		prettifiedEvent = &PresenceSharedUpdatedEvent{event}
	case *gchatproto.EventBody_UserOwnershipUpdatedEvent:
		prettifiedEvent = &UserOwnershipUpdatedEvent{event}
	case *gchatproto.EventBody_GroupScopedCapabilitiesEvent:
		prettifiedEvent = &GroupScopedCapabilitiesEvent{event}
	case *gchatproto.EventBody_MeetEvent:
		prettifiedEvent = &MeetEvent{event}
	case *gchatproto.EventBody_GroupUnreadThreadStateUpdatedEvent:
		prettifiedEvent = &GroupUnreadThreadStateUpdatedEvent{event}
	case *gchatproto.EventBody_WebchannelCheckEvent:
		prettifiedEvent = &WebchannelCheckEvent{event}
	case *gchatproto.EventBody_RecurringDndSettingsUpdatedEvent:
		prettifiedEvent = &RecurringDndSettingsUpdatedEvent{event}
	case *gchatproto.EventBody_MessageLabelsUpdatedEvent:
		prettifiedEvent = &MessageLabelsUpdatedEvent{event}
	case *gchatproto.EventBody_MessageReactionsSummary:
		prettifiedEvent = &MessageReactionsSummary{event}
	case *gchatproto.EventBody_GroupDefaultSortOrderUpdatedEvent:
		prettifiedEvent = &GroupDefaultSortOrderUpdatedEvent{event}
	case *gchatproto.EventBody_TopicLabelEvent:
		prettifiedEvent = &TopicLabelEvent{event}
	case *gchatproto.EventBody_StringSortOrderUpdatedEvent:
		prettifiedEvent = &StringSortOrderUpdatedEvent{event}
	case *gchatproto.EventBody_MessageLabelEvent:
		prettifiedEvent = &MessageLabelEvent{event}
	case *gchatproto.EventBody_GroupLabelEvent:
		prettifiedEvent = &GroupLabelEvent{event}
	case *gchatproto.EventBody_RosterSectionEvent:
		prettifiedEvent = &RosterSectionEvent{event}
	case *gchatproto.EventBody_BadgeCountUpdatedEvent:
		prettifiedEvent = &BadgeCountUpdatedEvent{event}
	case *gchatproto.EventBody_WorldRefreshedEvent:
		prettifiedEvent = &WorldRefreshedEvent{event}
	case *gchatproto.EventBody_JoinRequestedEvent:
		prettifiedEvent = &JoinRequestedEvent{event}
	case *gchatproto.EventBody_GroupEntityEvent:
		prettifiedEvent = &GroupEntityEvent{event}
	case *gchatproto.EventBody_MessageDetectedIntentEvent:
		prettifiedEvent = &MessageDetectedIntentEvent{event}
	case *gchatproto.EventBody_TopicMetadataUpdatedEvent:
		prettifiedEvent = &TopicMetadataUpdatedEvent{event}
	case *gchatproto.EventBody_GroupReadStateUpdatedEvent:
		prettifiedEvent = &GroupReadStateUpdatedEvent{event}
	default:
		switch evBody.EventType {
		case gchatproto.EventType_Enum_EVENT_TYPE_SESSION_READY:
			prettifiedEvent = &SessionReadyEvent{}
		default:
			prettifiedEvent = nil
		}
	}
	return prettifiedEvent
}
