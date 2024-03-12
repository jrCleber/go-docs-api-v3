package messaging

type Events string

const (
	APP_STATE                 Events = "app.state"
	APP_STATE_SYNC_COMPLETE   Events = "app.state:sync-complete"
	ARCHIVE_CHAT              Events = "archive.chat"
	BLOCK_LIST                Events = "block.list"
	BUSINESS_NAME             Events = "profile.pushName:business"
	CALL_ACCEPT               Events = "call.accept"
	CALL_OFFER                Events = "call.offer"
	CALL_OFFER_NOTICE         Events = "call.offer:notice"
	CALL_PRE_ACCEPT           Events = "call.pre:accept"
	CALL_RELAY_LATENCY        Events = "call.relay:latency"
	CALL_TERMINATE            Events = "call.terminate"
	CALL_TRANSPORT            Events = "call.transport"
	CHAT_CLEAR                Events = "chat.clear"
	CHAT_DELETE               Events = "chat.delete"
	CHAT_PRESENCE             Events = "chat.presence"
	CHAT_MUTE                 Events = "chat.mute"
	CLIENT_OUTDATED           Events = "client.outdated"
	CONNECT_FAILURE           Events = "connect.failure"
	CONNECTED                 Events = "whatsapp.connected"
	CONTACT                   Events = "contact.upsert"
	DELETE_FOR_ME             Events = "delete.for:me"
	DISCONNECTED              Events = "websocket.disconnected"
	INSTANCE_ERROR            Events = "instance.error"
	GROUP_INFO                Events = "group.info"
	HiSTORY_SYNC              Events = "history.sync"
	IDENTITY_CHANGE           Events = "identity.change"
	INSTANCE_STATUS           Events = "instance.status"
	JOINED_GROUP              Events = "joined.group"
	KEEP_ALIVE_RESTORED       Events = "keep.alive:restored"
	KEEP_ALIVE_TIMEOUT        Events = "keep.alive:timeout"
	LABEL_ASSOCIATION_CHAT    Events = "label.association:chat"
	LABEL_ASSOCIATION_MESSAGE Events = "label.association:message"
	LABEL_EDIT                Events = "label.edit"
	LOGGED_OUT                Events = "logged.out"
	MARK_CHAT_READ            Events = "mark.chat:read"
	MEDIA_RETRY               Events = "media.retry"
	MEDIA_RETRY_ERROR         Events = "media.retry:error"
	NEW_MESSAGE               Events = "new.message"
	NEWS_LATTER_JOIN          Events = "news.latter:join"
	NEWS_LATTER_LEAVE         Events = "news.latter:leave"
	NEWS_LATTER_LIVE_UPDATE   Events = "news.latter:live-update"
	NEWS_LATTER_MESSAGE_META  Events = "news.latter:message-meta"
	NEWS_LATTER_MUTE_CHANGE   Events = "news.latter:mute-change"
	OFFLINE_SYNC_COMPLETED    Events = "offline.sync:completed"
	OFFLINE_SYNC_PREVIEW      Events = "offline.sync:preview"
	DEVICE_PARING             Events = "device.paring"
	PICTURE                   Events = "profile.picture"
	PINNED_CHAT               Events = "pinned.chat"
	PRESENCE                  Events = "presence.update"
	PRIVACY_SETTINGS          Events = "privacy.settings"
	PUSH_NAME                 Events = "profile.pushName"
	PUSH_NAME_SELF            Events = "profile.pushName:self"
	QR_CODE                   Events = "qrcode.update"
	RECEIPT                   Events = "message.update"
	MESSAGE_STAR              Events = "message.star"
	STATUS_BROADCAST          Events = "status.broadcast"
	SEND_MESSAGE              Events = "send.message"
	CONNECTION_STREAM         Events = "connection.stream"
	TEMPORARY_BAN             Events = "temporary.ban"
	UNARCHIVE_CHAT_SETTINGS   Events = "chat.settings:archived"
	UNDECRYPTABLE_MESSAGE     Events = "undecryptable.message"
	UNKNOWN_CALL              Events = "unknown.call"
	USER_STATUS_MUTE          Events = "user.status:mute"
)

func (m *Amqp) SetQueues() {
	m.SetupExchangesAndQueues([]string{
		string(APP_STATE),
		string(APP_STATE_SYNC_COMPLETE),
		string(ARCHIVE_CHAT),
		string(BLOCK_LIST),
		string(BUSINESS_NAME),
		string(CALL_ACCEPT),
		string(CALL_OFFER),
		string(CALL_OFFER_NOTICE),
		string(CALL_PRE_ACCEPT),
		string(CALL_RELAY_LATENCY),
		string(CALL_TERMINATE),
		string(CALL_TRANSPORT),
		string(CHAT_PRESENCE),
		string(CHAT_CLEAR),
		string(CHAT_DELETE),
		string(CHAT_MUTE),
		string(CLIENT_OUTDATED),
		string(CONNECT_FAILURE),
		string(CONNECTED),
		string(CONTACT),
		string(CHAT_DELETE),
		string(DELETE_FOR_ME),
		string(DISCONNECTED),
		string(GROUP_INFO),
		string(HiSTORY_SYNC),
		string(IDENTITY_CHANGE),
		string(INSTANCE_STATUS),
		// string(JOINED_GROUP),
		string(KEEP_ALIVE_RESTORED),
		string(KEEP_ALIVE_TIMEOUT),
		string(LABEL_ASSOCIATION_CHAT),
		string(LABEL_ASSOCIATION_MESSAGE),
		string(LABEL_EDIT),
		string(LOGGED_OUT),
		string(MARK_CHAT_READ),
		string(MEDIA_RETRY),
		string(MEDIA_RETRY_ERROR),
		string(NEW_MESSAGE),
		string(MESSAGE_STAR),
		string(CHAT_MUTE),
		// string(NEWS_LATTER_JOIN),
		// string(NEWS_LATTER_LEAVE),
		// string(NEWS_LATTER_LIVE_UPDATE),
		// string(NEWS_LATTER_MESSAGE_META),
		// string(NEWS_LATTER_MUTE_CHANGE),
		string(OFFLINE_SYNC_COMPLETED),
		string(OFFLINE_SYNC_PREVIEW),
		string(DEVICE_PARING),
		string(PICTURE),
		string(PINNED_CHAT),
		string(QR_CODE),
		string(RECEIPT),
		string(SEND_MESSAGE),
		string(STATUS_BROADCAST),
		string(CONNECTION_STREAM),
		string(TEMPORARY_BAN),
		string(UNARCHIVE_CHAT_SETTINGS),
		string(UNDECRYPTABLE_MESSAGE),
		string(UNKNOWN_CALL),
		string(USER_STATUS_MUTE),
	})
}
