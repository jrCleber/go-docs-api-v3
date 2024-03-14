package whatsapp

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"time"

	"codechat.dev/pkg/config"
	"codechat.dev/pkg/messaging"
	"codechat.dev/pkg/utils"
	"github.com/google/uuid"
	"github.com/mdp/qrterminal"
	"github.com/sirupsen/logrus"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type StateEnum string

const (
	ACTIVE   StateEnum = "active"
	INACTIVE StateEnum = "inactive"
)

type WhatsApp struct {
	ID             string     `json:"whatsappId,omitempty"`
	Number         string     `json:"number,omitempty"`
	PictureUrl     string     `json:"pictureUrl,omitempty"`
	PushName       string     `json:"pushName,omitempty"`
	LastConnection *time.Time `json:"lastConnection,omitempty"`
}

type QrCode struct {
	Code       string `json:"code,omitempty"`
	Base64     string `json:"base64,omitempty"`
	Connection string `json:"connection,omitempty"`
}

type StatusEnum string

const (
	Created   StatusEnum = "created"
	Available StatusEnum = "available"
	Booting   StatusEnum = "booting"
	Waiting   StatusEnum = "waiting"
	Deleted   StatusEnum = "deleted"
)

type ConnectionStatusEnum string

const (
	OPEN       ConnectionStatusEnum = "open"
	CLOSE      ConnectionStatusEnum = "close"
	REFUSED    ConnectionStatusEnum = "refused"
	CONNECTING ConnectionStatusEnum = "connecting"
)

type Instance struct {
	ID            string               `json:"instanceId"`
	Name          string               `json:"name"`
	State         StateEnum            `json:"state,omitempty"`
	Apikey        *string              `json:"apikey,omitempty"`
	Status        StatusEnum           `json:"status,omitempty"`
	Connection    ConnectionStatusEnum `json:"connection"`
	ExternalId    string               `json:"externalId,omitempty"`
	QR            *QrCode              `json:"-"`
	CreatedAt     time.Time            `json:"createdAt"`
	UpdateAt      *time.Time           `json:"updateAt,omitempty"`
	DeletedAt     *time.Time           `json:"deletedAt,omitempty"`
	WhatsApp      *WhatsApp            `json:"WhatsApp,omitempty"`
	Client        *whatsmeow.Client    `json:"-"`
	Logger        *logrus.Entry        `json:"-"`
	Messaging     *messaging.Amqp      `json:"-"`
	Store         *Store               `json:"-"`
}

type SendMessage struct {
	Event    messaging.Events
	Instance *Instance
	Data     any
	GlobalWebhook string
}

type ConnectionUpdate struct {
	Status       ConnectionStatusEnum        `json:"status"`
	StatusReason events.ConnectFailureReason `json:"statusReason"`
	Message      string                      `json:"Message,omitempty"`
	Timestamp    time.Time                   `json:"timestamp"`
}

func NewInstance(
	id, name, externalId, phoneNumber, apikey string,
	state StateEnum, messaging *messaging.Amqp,
	store *Store, containerName string,
) *Instance {
	if state == "" {
		state = ACTIVE
	}
	if id == "" {
		id = uuid.NewString()
	}

	if apikey == "" {
		apikey = strings.ToUpper(uuid.NewString())
	}

	instance := Instance{
		Name:       name,
		ID:         id,
		State:      state,
		Apikey:     &apikey,
		Status:     Waiting,
		ExternalId: externalId,
		Connection: CLOSE,
		CreatedAt:  time.Now(),
		UpdateAt:   nil,
		DeletedAt:  nil,
		WhatsApp: &WhatsApp{
			Number: phoneNumber,
		},
		Messaging:     messaging,
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.WithFields(logrus.Fields{"instance": name, "id": instance.ID})

	instance.Logger = logger.WithFields(logrus.Fields{"instance": name, "id": instance.ID})

	return &instance
}

func (i *Instance) ConnectionsStatus() ConnectionStatusEnum {
	if i.Client.IsLoggedIn() {
		return OPEN
	}
	return CLOSE
}

func (i *Instance) GetInstance() *Instance {
	i.Connection = i.ConnectionsStatus()
	i.WhatsApp.Number = i.Client.Store.ID.User
	i.WhatsApp.PushName = i.Client.Store.PushName
	rawJid := utils.FormatJid(i.WhatsApp.Number)
	jid, _ := types.ParseJID(rawJid)
	pic, err := i.Client.GetProfilePictureInfo(jid, &whatsmeow.GetProfilePictureParams{
		Preview: false,
	})
	if err != nil {
		i.Logger.Errorln("Unable to retrieve profile image for the connected number: ", i.WhatsApp.Number, err)
	} else {
		i.WhatsApp.PictureUrl = pic.URL
	}
	i.WhatsApp.LastConnection = &i.Client.LastSuccessfulConnect

	return i
}

func (i *Instance) connectionUpdate() error {
	if i.Client.Store.ID == nil {
		qrChan, _ := i.Client.GetQRChannel(context.Background())
		err := i.Client.Connect()
		if err != nil {
			return err
		}

		go func() {
			for ev := range qrChan {
				if ev.Event == "code" {
					qr, err := qrcode.Encode(ev.Code, qrcode.High, 256)
					if err != nil {
						i.Logger.WithFields(logrus.Fields{"event": "qrcode"}).Error("Failed to generate the qr code: ", err)
					}

					i.QR = &QrCode{
						Code:       ev.Code,
						Base64:     base64.StdEncoding.EncodeToString(qr),
						Connection: string(CONNECTING),
					}

					i.Messaging.SendMessage(string(messaging.QR_CODE), PreparedMessage(messaging.QR_CODE, i, i.QR))

					qrterminal.GenerateHalfBlock(i.QR.Code, qrterminal.L, os.Stdout)

				} else {
					i.Logger.WithFields(logrus.Fields{"event": "qrcode"}).Info("Login")
				}
			}
		}()
	} else {
		err := i.Client.Connect()
		if err != nil {
			return err
		}
		i.QR = &QrCode{Connection: string(OPEN)}
		i.WhatsApp.Number = i.Client.Store.ID.User
	}

	return nil
}

func PreparedMessage(evt messaging.Events, instance *Instance, value any) *SendMessage {
	return &SendMessage{
		Event:    evt,
		Instance: instance,
		Data:     value,
		GlobalWebhook: config.GlobalWebhook,
	}
}

func (i *Instance) NewConnection() error {
	if i.Status == Booting {
		return errors.New(utils.StringJoin("", "Instance is not fully loaded - status: ", string(i.Status)))
	}

	device, err := utils.NewDevice()
	if err != nil {
		return err
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(device, clientLog)

	i.Client = client

	err = i.connectionUpdate()
	if err != nil {
		return err
	}

	client.AddEventHandler(func(evt any) {
		i.handlers(evt)
	})

	i.Messaging.SendMessage(
		string(messaging.INSTANCE_STATUS),
		PreparedMessage(messaging.INSTANCE_STATUS, i, map[string]StatusEnum{
			"Status": Available,
		}))

	return nil
}

func (i *Instance) Connect() error {
	deviceStore, err := utils.LoadDeviceWa(&i.WhatsApp.Number)
	if err != nil {
		return err
	}

	clientLog := waLog.Stdout("Client", "ERROR", true)
	client := whatsmeow.NewClient(deviceStore, clientLog)

	if client.Store == nil {
		return errors.New("no Store")
	}

	i.Client = client

	err = i.connectionUpdate()
	if err != nil {
		return err
	}

	client.AddEventHandler(func(evt any) {
		i.handlers(evt)
	})

	i.Messaging.SendMessage(
		string(messaging.INSTANCE_STATUS),
		PreparedMessage(messaging.INSTANCE_STATUS, i, map[string]StatusEnum{
			"Status": Available,
		}))

	return nil
}

func (i *Instance) QrCode() *QrCode {
	return i.QR
}

func (i *Instance) handlers(evt interface{}) {
	client := i.Messaging

	switch value := evt.(type) {
	case *events.AppState:
		client.SendMessage(string(messaging.APP_STATE), PreparedMessage(messaging.APP_STATE, i, value))
		return
	case *events.AppStateSyncComplete:
		client.SendMessage(string(messaging.APP_STATE_SYNC_COMPLETE), PreparedMessage(messaging.APP_STATE_SYNC_COMPLETE, i, value))
		return
	case *events.Archive:
		client.SendMessage(string(messaging.ARCHIVE_CHAT), PreparedMessage(messaging.ARCHIVE_CHAT, i, value))
		return
	case *events.Blocklist:
		client.SendMessage(string(messaging.BLOCK_LIST), PreparedMessage(messaging.BLOCK_LIST, i, value))
		return
	case *events.BusinessName:
		client.SendMessage(string(messaging.BUSINESS_NAME), PreparedMessage(messaging.BUSINESS_NAME, i, value))
		return
	case *events.CallAccept:
		client.SendMessage(string(messaging.CALL_ACCEPT), PreparedMessage(messaging.CALL_ACCEPT, i, value))
		return
	case *events.CallOffer:
		client.SendMessage(string(messaging.CALL_OFFER), PreparedMessage(messaging.CALL_OFFER, i, value))
		return
	case *events.CallOfferNotice:
		client.SendMessage(string(messaging.CALL_OFFER_NOTICE), PreparedMessage(messaging.CALL_OFFER_NOTICE, i, value))
		return
	case *events.CallRelayLatency:
		client.SendMessage(string(messaging.CALL_RELAY_LATENCY), PreparedMessage(messaging.CALL_RELAY_LATENCY, i, value))
		return
	case *events.CallTerminate:
		client.SendMessage(string(messaging.CALL_TERMINATE), PreparedMessage(messaging.CALL_TERMINATE, i, value))
		return
	case *events.ChatPresence:
		client.SendMessage(string(messaging.PRESENCE), PreparedMessage(messaging.PRESENCE, i, value))
		return
	case *events.ClearChat:
		client.SendMessage(string(messaging.CHAT_CLEAR), PreparedMessage(messaging.CHAT_CLEAR, i, value))
		return
	case *events.ClientOutdated:
		client.SendMessage(string(messaging.CLIENT_OUTDATED), PreparedMessage(messaging.CLIENT_OUTDATED, i, ConnectionUpdate{
			Status:       REFUSED,
			StatusReason: 409,
			Message:      "client user agent was rejected",
			Timestamp:    time.Now(),
		}))
		return
	case *events.ConnectFailure:
		i.Status = Waiting
		client.SendMessage(string(messaging.CONNECT_FAILURE), PreparedMessage(messaging.CONNECT_FAILURE, i, ConnectionUpdate{
			Status:       REFUSED,
			StatusReason: value.Reason,
			Message:      value.Message,
			Timestamp:    time.Now(),
		}))
		return
	case *events.Connected:
		i.Status = Available
		i.Connection = OPEN
		i.WhatsApp.Number = strings.Split(i.Client.Store.ID.String(), ":")[0]
		i.WhatsApp.LastConnection = &i.Client.LastSuccessfulConnect
		client.SendMessage(string(messaging.CONNECTED), PreparedMessage(messaging.CONNECTED, i, ConnectionUpdate{
			Status:       i.Connection,
			StatusReason: 200,
			Message:      "connected successfully",
			Timestamp:    time.Now(),
		}))
		i.Store.Update(i.ID, &Instance{
			WhatsApp: &WhatsApp{
				Number: i.Client.Store.ID.User,
			},
		})
		return
	case *events.Contact:
		client.SendMessage(string(messaging.CONTACT), PreparedMessage(messaging.CONTACT, i, value))
		return
	case *events.DeleteChat:
		client.SendMessage(string(messaging.CHAT_DELETE), PreparedMessage(messaging.CHAT_DELETE, i, value))
		return
	case *events.DeleteForMe:
		client.SendMessage(string(messaging.DELETE_FOR_ME), PreparedMessage(messaging.DELETE_FOR_ME, i, value))
		return
	case *events.Disconnected:
		client.SendMessage(string(messaging.DISCONNECTED), PreparedMessage(messaging.DISCONNECTED, i, ConnectionUpdate{
			Status:       i.Connection,
			StatusReason: events.ConnectFailureLoggedOut,
			Timestamp:    time.Now(),
		}))
		return
	case *events.GroupInfo:
		client.SendMessage(string(messaging.GROUP_INFO), PreparedMessage(messaging.GROUP_INFO, i, value))
		return
	case *events.HistorySync:
		client.SendMessage(string(messaging.HiSTORY_SYNC), PreparedMessage(messaging.HiSTORY_SYNC, i, value.Data))
		return
	case *events.IdentityChange:
		client.SendMessage(string(messaging.IDENTITY_CHANGE), PreparedMessage(messaging.IDENTITY_CHANGE, i, value))
		return
	// case *events.JoinedGroup:
	// 	client.SendMessage(string(messaging.JOINED_GROUP), PreparedMessage(messaging.JOINED_GROUP, i, value))
	// 	return
	case *events.KeepAliveRestored:
		client.SendMessage(string(messaging.KEEP_ALIVE_RESTORED), PreparedMessage(messaging.KEEP_ALIVE_RESTORED, i, map[string]any{
			"Type":      "restored",
			"Timestamp": time.Now(),
		}))
		return
	case *events.KeepAliveTimeout:
		client.SendMessage(string(messaging.KEEP_ALIVE_TIMEOUT), PreparedMessage(messaging.KEEP_ALIVE_TIMEOUT, i, map[string]any{
			"Type":      "timeout",
			"Timestamp": time.Now(),
		}))
		return
	case *events.LabelAssociationChat:
		client.SendMessage(string(messaging.LABEL_ASSOCIATION_CHAT), PreparedMessage(messaging.LABEL_ASSOCIATION_CHAT, i, value))
	case *events.LabelAssociationMessage:
		client.SendMessage(string(messaging.LABEL_ASSOCIATION_MESSAGE), PreparedMessage(messaging.LABEL_ASSOCIATION_MESSAGE, i, value))
		return
	case *events.LabelEdit:
		client.SendMessage(string(messaging.LABEL_EDIT), PreparedMessage(messaging.LABEL_EDIT, i, value))
		return
	case *events.LoggedOut:
		i.Status = Waiting
		i.Connection = CLOSE
		client.SendMessage(string(messaging.LOGGED_OUT), PreparedMessage(messaging.LOGGED_OUT, i, map[string]any{
			"OnConnect":   false,
			"Reason":      value.Reason,
			"Description": value.Reason.String(),
		}))
		i.Store.Update(i.ID, &Instance{
			Connection: i.Connection,
			Status:     i.Status,
		})
	case *events.MarkChatAsRead:
		client.SendMessage(string(messaging.MARK_CHAT_READ), PreparedMessage(messaging.MARK_CHAT_READ, i, value))
		return
	case *events.MediaRetry:
		client.SendMessage(string(messaging.MEDIA_RETRY), PreparedMessage(messaging.MEDIA_RETRY, i, value))
		return
	case *events.MediaRetryError:
		client.SendMessage(string(messaging.MEDIA_RETRY_ERROR), PreparedMessage(messaging.MEDIA_RETRY_ERROR, i, value))
		return
	case *events.Message:
		event := messaging.NEW_MESSAGE
		if value.Info.Chat.String() == "status@broadcast" {
			event = messaging.STATUS_BROADCAST
		}
		value.Info.Sender.RawAgent = 0
		value.Info.Sender.Device = 0
		value.RawMessage = nil
		client.SendMessage(string(event), PreparedMessage(event, i, value))
		return
	case *events.Mute:
		client.SendMessage(string(messaging.CHAT_MUTE), PreparedMessage(messaging.CHAT_MUTE, i, value))
		return
	// case *events.NewsletterJoin:
	// 	client.SendMessage(string(messaging.NEWS_LATTER_JOIN), PreparedMessage(messaging.NEWS_LATTER_JOIN, i, value))
	// 	return
	// case *events.NewsletterLeave:
	// 	client.SendMessage(string(messaging.NEWS_LATTER_LEAVE), PreparedMessage(messaging.NEWS_LATTER_LEAVE, i, value))
	// 	return
	// case *events.NewsletterLiveUpdate:
	// 	client.SendMessage(string(messaging.NEWS_LATTER_LIVE_UPDATE), PreparedMessage(messaging.NEWS_LATTER_LIVE_UPDATE, i, value))
	// 	return
	// case *events.NewsletterMessageMeta:
	// 	client.SendMessage(string(messaging.NEWS_LATTER_MESSAGE_META), PreparedMessage(messaging.NEWS_LATTER_MESSAGE_META, i, value))
	// 	return
	// case *events.NewsletterMuteChange:
	// 	client.SendMessage(string(messaging.NEWS_LATTER_MUTE_CHANGE), PreparedMessage(messaging.NEWS_LATTER_MUTE_CHANGE, i, value))
	// 	return
	case *events.OfflineSyncCompleted:
		client.SendMessage(string(messaging.OFFLINE_SYNC_COMPLETED), PreparedMessage(messaging.OFFLINE_SYNC_COMPLETED, i, value))
		return
	case *events.OfflineSyncPreview:
		client.SendMessage(string(messaging.OFFLINE_SYNC_PREVIEW), PreparedMessage(messaging.OFFLINE_SYNC_PREVIEW, i, value))
		return
	case *events.PairError:
		client.SendMessage(string(messaging.DEVICE_PARING), PreparedMessage(messaging.DEVICE_PARING, i, map[string]any{
			"Type":         "error",
			"ID":           value.ID,
			"BusinessName": value.BusinessName,
			"Error":        value.Error,
		}))
		return
	case *events.PairSuccess:
		client.SendMessage(string(messaging.DEVICE_PARING), PreparedMessage(messaging.DEVICE_PARING, i, map[string]any{
			"Type":         "success",
			"ID":           value.ID,
			"BusinessName": value.BusinessName,
		}))
		return
	case *events.Picture:
		picture, _ := i.Client.GetProfilePictureInfo(value.JID, &whatsmeow.GetProfilePictureParams{
			Preview:    false,
			ExistingID: value.PictureID,
		})
		client.SendMessage(string(messaging.PICTURE), PreparedMessage(messaging.PICTURE, i, map[string]any{
			"JID":         value.JID,
			"Author":      value.Author,
			"Timestamp":   value.Timestamp,
			"Remove":      value.Remove,
			"PictureID":   value.PictureID,
			"PictureInfo": picture,
		}))
		return
	case *events.Pin:
		client.SendMessage(string(messaging.PINNED_CHAT), PreparedMessage(messaging.PINNED_CHAT, i, value))
		return
	case *events.Presence:
		client.SendMessage(string(messaging.PRESENCE), PreparedMessage(messaging.PRESENCE, i, value))
		return
	case *events.PrivacySettings:
		client.SendMessage(string(messaging.PRIVACY_SETTINGS), PreparedMessage(messaging.PRIVACY_SETTINGS, i, value))
	case *events.PushName:
		client.SendMessage(string(messaging.PUSH_NAME), PreparedMessage(messaging.PUSH_NAME, i, value))
		return
	case *events.PushNameSetting:
		client.SendMessage(string(messaging.PUSH_NAME_SELF), PreparedMessage(messaging.PUSH_NAME_SELF, i, value))
		return
	// case *events.QR:
	// 	client.SendMessage(string(messaging. QR_CODE), PreparedMessage(messaging. QR_CODE, i, value))
	// 	return
	case *events.Receipt:
		if value.Type == "" {
			value.Type = "delivery-ack"
		}
		client.SendMessage(string(messaging.RECEIPT), PreparedMessage(messaging.RECEIPT, i, value))
		return
	case *events.Star:
		client.SendMessage(string(messaging.MESSAGE_STAR), PreparedMessage(messaging.MESSAGE_STAR, i, value))
		return
	case *events.StreamError:
		client.SendMessage(string(messaging.CONNECTION_STREAM), PreparedMessage(messaging.CONNECTION_STREAM, i, map[string]any{
			"Type": "error",
			"Code": value.Code,
			"Info": value.Raw,
		}))
		return
	case *events.StreamReplaced:
		client.SendMessage(string(messaging.CONNECTION_STREAM), PreparedMessage(messaging.CONNECTION_STREAM, i, map[string]any{
			"Type": "replaced",
			"Code": value.PermanentDisconnectDescription(),
			"Info": map[string]string{"Content": "conflict"},
		}))
		return
	case *events.TemporaryBan:
		client.SendMessage(string(messaging.TEMPORARY_BAN), PreparedMessage(messaging.TEMPORARY_BAN, i, map[string]any{
			"Code":    value.Code,
			"Expire":  value.Expire,
			"Message": value.Code.String(),
		}))
		return
	case *events.UnarchiveChatsSetting:
		client.SendMessage(string(messaging.UNARCHIVE_CHAT_SETTINGS), PreparedMessage(messaging.UNARCHIVE_CHAT_SETTINGS, i, value))
		return
	case *events.UndecryptableMessage:
		client.SendMessage(string(messaging.UNDECRYPTABLE_MESSAGE), PreparedMessage(messaging.UNDECRYPTABLE_MESSAGE, i, value))
		return
	case *events.UnknownCallEvent:
		client.SendMessage(string(messaging.UNKNOWN_CALL), PreparedMessage(messaging.UNKNOWN_CALL, i, value))
		return
	case *events.UserStatusMute:
		client.SendMessage(string(messaging.USER_STATUS_MUTE), PreparedMessage(messaging.USER_STATUS_MUTE, i, value))
		return
	}
}
