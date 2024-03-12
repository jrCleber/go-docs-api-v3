package contract

import (
	"go.mau.fi/whatsmeow/binary/proto"
)

type Message struct {
	Conversation                               *string                                 `json:"conversation,omitempty"`
	SenderKeyDistributionMessage               *proto.SenderKeyDistributionMessage     `json:"senderKeyDistributionMessage,omitempty"`
	ImageMessage                               *proto.ImageMessage                     `json:"imageMessage,omitempty"`
	ContactMessage                             *proto.ContactMessage                   `json:"contactMessage,omitempty"`
	LocationMessage                            *proto.LocationMessage                  `json:"locationMessage,omitempty"`
	ExtendedTextMessage                        *proto.ExtendedTextMessage              `json:"extendedTextMessage,omitempty"`
	DocumentMessage                            *proto.DocumentMessage                  `json:"documentMessage,omitempty"`
	AudioMessage                               *proto.AudioMessage                     `json:"audioMessage,omitempty"`
	VideoMessage                               *proto.VideoMessage                     `json:"videoMessage,omitempty"`
	Call                                       *proto.Call                             `json:"call,omitempty"`
	Chat                                       *proto.Chat                             `json:"chat,omitempty"`
	ProtocolMessage                            *proto.ProtocolMessage                  `json:"protocolMessage,omitempty"`
	ContactsArrayMessage                       *proto.ContactsArrayMessage             `json:"contactsArrayMessage,omitempty"`
	HighlyStructuredMessage                    *proto.HighlyStructuredMessage          `json:"highlyStructuredMessage,omitempty"`
	FastRatchetKeySenderKeyDistributionMessage *proto.SenderKeyDistributionMessage     `json:"fastRatchetKeySenderKeyDistributionMessage,omitempty"`
	SendPaymentMessage                         *proto.SendPaymentMessage               `json:"sendPaymentMessage,omitempty"`
	LiveLocationMessage                        *proto.LiveLocationMessage              `json:"liveLocationMessage,omitempty"`
	RequestPaymentMessage                      *proto.RequestPaymentMessage            `json:"requestPaymentMessage,omitempty"`
	DeclinePaymentRequestMessage               *proto.DeclinePaymentRequestMessage     `json:"declinePaymentRequestMessage,omitempty"`
	CancelPaymentRequestMessage                *proto.CancelPaymentRequestMessage      `json:"cancelPaymentRequestMessage,omitempty"`
	TemplateMessage                            *proto.TemplateMessage                  `json:"templateMessage,omitempty"`
	StickerMessage                             *proto.StickerMessage                   `json:"stickerMessage,omitempty"`
	GroupInviteMessage                         *proto.GroupInviteMessage               `json:"groupInviteMessage,omitempty"`
	TemplateButtonReplyMessage                 *proto.TemplateButtonReplyMessage       `json:"templateButtonReplyMessage,omitempty"`
	ProductMessage                             *proto.ProductMessage                   `json:"productMessage,omitempty"`
	DeviceSentMessage                          *proto.DeviceSentMessage                `json:"deviceSentMessage,omitempty"`
	MessageContextInfo                         *proto.MessageContextInfo               `json:"messageContextInfo,omitempty"`
	ListMessage                                *proto.ListMessage                      `json:"listMessage,omitempty"`
	ViewOnceMessage                            *proto.FutureProofMessage               `json:"viewOnceMessage,omitempty"`
	OrderMessage                               *proto.OrderMessage                     `json:"orderMessage,omitempty"`
	ListResponseMessage                        *proto.ListResponseMessage              `json:"listResponseMessage,omitempty"`
	EphemeralMessage                           *proto.FutureProofMessage               `json:"ephemeralMessage,omitempty"`
	InvoiceMessage                             *proto.InvoiceMessage                   `json:"invoiceMessage,omitempty"`
	ButtonsMessage                             *proto.ButtonsMessage                   `json:"buttonsMessage,omitempty"`
	ButtonsResponseMessage                     *proto.ButtonsResponseMessage           `json:"buttonsResponseMessage,omitempty"`
	PaymentInviteMessage                       *proto.PaymentInviteMessage             `json:"paymentInviteMessage,omitempty"`
	InteractiveMessage                         *proto.InteractiveMessage               `json:"interactiveMessage,omitempty"`
	ReactionMessage                            *proto.ReactionMessage                  `json:"reactionMessage,omitempty"`
	StickerSyncRmrMessage                      *proto.StickerSyncRMRMessage            `json:"stickerSyncRmrMessage,omitempty"`
	InteractiveResponseMessage                 *proto.InteractiveResponseMessage       `json:"interactiveResponseMessage,omitempty"`
	PollCreationMessage                        *proto.PollCreationMessage              `json:"pollCreationMessage,omitempty"`
	PollUpdateMessage                          *proto.PollUpdateMessage                `json:"pollUpdateMessage,omitempty"`
	KeepInChatMessage                          *proto.KeepInChatMessage                `json:"keepInChatMessage,omitempty"`
	DocumentWithCaptionMessage                 *proto.FutureProofMessage               `json:"documentWithCaptionMessage,omitempty"`
	RequestPhoneNumberMessage                  *proto.RequestPhoneNumberMessage        `json:"requestPhoneNumberMessage,omitempty"`
	ViewOnceMessageV2                          *proto.FutureProofMessage               `json:"viewOnceMessageV2,omitempty"`
	EncReactionMessage                         *proto.EncReactionMessage               `json:"encReactionMessage,omitempty"`
	EditedMessage                              *proto.FutureProofMessage               `json:"editedMessage,omitempty"`
	ViewOnceMessageV2Extension                 *proto.FutureProofMessage               `json:"viewOnceMessageV2Extension,omitempty"`
	PollCreationMessageV2                      *proto.PollCreationMessage              `json:"pollCreationMessageV2,omitempty"`
	ScheduledCallCreationMessage               *proto.ScheduledCallCreationMessage     `json:"scheduledCallCreationMessage,omitempty"`
	GroupMentionedMessage                      *proto.FutureProofMessage               `json:"groupMentionedMessage,omitempty"`
	PinInChatMessage                           *proto.PinInChatMessage                 `json:"pinInChatMessage,omitempty"`
	PollCreationMessageV3                      *proto.PollCreationMessage              `json:"pollCreationMessageV3,omitempty"`
	ScheduledCallEditMessage                   *proto.ScheduledCallEditMessage         `json:"scheduledCallEditMessage,omitempty"`
	PtvMessage                                 *proto.VideoMessage                     `json:"ptvMessage,omitempty"`
	BotInvokeMessage                           *proto.FutureProofMessage               `json:"botInvokeMessage,omitempty"`
	CallLogMesssage                            *proto.CallLogMessage                   `json:"callLogMesssage,omitempty"`
	MessageHistoryBundle                       *proto.MessageHistoryBundle             `json:"messageHistoryBundle,omitempty"`
	EncCommentMessage                          *proto.EncCommentMessage                `json:"encCommentMessage,omitempty"`
	BcallMessage                               *proto.BCallMessage                     `json:"bcallMessage,omitempty"`
	LottieStickerMessage                       *proto.FutureProofMessage               `json:"lottieStickerMessage,omitempty"`
	EventMessage                               *proto.EventMessage                     `json:"eventMessage,omitempty"`
	EncEventResponseMessage                    *proto.EncEventResponseMessage          `json:"encEventResponseMessage,omitempty"`
	CommentMessage                             *proto.CommentMessage                   `json:"commentMessage,omitempty"`
	NewsletterAdminInviteMessage               *proto.NewsletterAdminInviteMessage     `json:"newsletterAdminInviteMessage,omitempty"`
	ExtendedTextMessageWithParentKey           *proto.ExtendedTextMessageWithParentKey `json:"extendedTextMessageWithParentKey,omitempty"`
	PlaceholderMessage                         *proto.PlaceholderMessage               `json:"placeholderMessage,omitempty"`
	SecretEncryptedMessage                     *proto.SecretEncryptedMessage           `json:"secretEncryptedMessage,omitempty"`
}

type Group struct {
	HiddenMention bool `json:"hiddenMention"`
}

type RecipientParam struct {
	Recipient string `validate:"required" json:"recipient"`
}

type Quoted struct {
	MessageId string         `validate:"required" json:"messageId"`
	Sender    string         `validate:"required" json:"sender"`
	Message   *proto.Message `validate:"required" json:"message"`
}

type Options struct {
	MessageID          string      `json:"messageId"`
	Delay              int         `validate:"gte=0" json:"delay"`
	Presence           string      `validate:"oneof=composing recording none" json:"presence"`
	QuotedMessage      interface{} `json:"quotedMessage"`
	Group              Group       `json:"groupMention"`
	ExternalAttributes string      `json:"externalAttributes"`
}

type Base struct {
	RecipientParam
	Options Options `json:"options"`
}

type Text struct {
	Text string `validate:"required" json:"text"`
}

type TextMessage struct {
	Base
	Message Text `validate:"required" json:"textMessage"`
}

type Link struct {
	Link        string `validate:"required" json:"link"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Text        string `json:"text"`
}
type LinkMessage struct {
	Base
	Message Link `validate:"required" json:"linkMessage"`
}

type AttrMedia struct {
	MediaType   string `validate:"required,oneof=image video sticker document audio ptv" json:"mediatype"`
	Caption     string `json:"caption"`
	Filename    string `json:"filename"`
	GifPlayback bool   `json:"isGif"`
}

type UrlMedia struct {
	Url string `validate:"required,isUrl" json:"url"`
	AttrMedia
}
type MediaMessage struct {
	Base
	Message UrlMedia `validate:"required" json:"mediaMessage"`
}

type AudioMessage struct {
	Base
	Message UrlMedia `validate:"required" json:"audioMessage"`
}

type PtvMessage struct {
	Base
	Message UrlMedia `validate:"required" json:"ptvMessage"`
}

type Location struct {
	Name      string  `json:"name"`
	Address   string  `json:"address"`
	Comment   string  `json:"comment"`
	Latitude  float64 `validate:"required" json:"latitude"`
	Longitude float64 `validate:"required" json:"longitude"`
}

type LocationMessage struct {
	Base
	Message Location `validate:"required" json:"locationMessage"`
}

type Contact struct {
	FullName    string `validate:"required" json:"fullName"`
	RawNumber   string `validate:"required" json:"rawNumber"`
	PhoneNumber string `validate:"required" json:"phoneNumber"`
}

type ContactMessage struct {
	Base
	Message []Contact `validate:"required,gt=0,dive,uniqueStruct" json:"contactMessage"`
}

type Reaction struct {
	Reaction string `validate:"required" json:"reaction"`
}

type MessageKey struct {
	ChatID    string `validate:"required" json:"chatId"`
	MessageID string `validate:"required" json:"messageId"`
	IsFromMe  bool   `validate:"required" json:"fromMe"`
}

type ReactionMessage struct {
	Base
	MessageKey MessageKey `validate:"required" json:"messageKey"`
	Message    Reaction   `validate:"required" json:"reactionMessage"`
}

type edit struct {
	NewContent string `validate:"required" json:"text"`
}

type EditMessage struct {
	RecipientParam
	Message edit `validate:"required" json:"editMessage"`
}

type Row struct {
	Title       string `validate:"required" json:"title"`
	Description string `json:"description"`
	RowId       string `validate:"required" json:"rowId"`
}

type Section struct {
	Title string `validate:"required" json:"title"`
	Rows  []Row  `validate:"required,gt=0,dive" json:"rows"`
}

type List struct {
	Title       string    `validate:"required" json:"title"`
	Description string    `json:"description"`
	ButtonText  string    `validate:"required" json:"buttonText"`
	FooterText  string    `json:"footerText"`
	Sections    []Section `validate:"required,gt=0,dive" json:"sections"`
}

type ListMessage struct {
	Base
	Message List `validate:"required" json:"listMessage"`
}

type Poll struct {
	Name                   string   `validate:"required" json:"name"`
	Options                []string `validate:"required,uniqueString" json:"options"`
	SelectableOptionsCount uint32   `json:"selectableOptionsCount "`
}

type PollMessage struct {
	Base
	Message Poll `validate:"required" json:"pollMessage"`
}
