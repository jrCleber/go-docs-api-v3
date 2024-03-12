package sendmessage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"codechat.dev/contract"
	"codechat.dev/internal/domain/instance"
	"codechat.dev/internal/whatsapp"
	"codechat.dev/pkg/messaging"
	"codechat.dev/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)

type Service struct {
	Logger    *logrus.Entry
	Manager   *instance.InstancesManager
	Messaging *messaging.Amqp
	Store     *whatsapp.Store
	Ctx       context.Context
}

func NewService(store *whatsapp.Store, manager *instance.InstancesManager, messaging *messaging.Amqp, ctx context.Context) *Service {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &Service{
		Logger:    logger.WithFields(logrus.Fields{"name": "sendmessage-service"}),
		Manager:   manager,
		Messaging: messaging,
		Store:     store,
		Ctx:       ctx,
	}
}

func setMentionedGroup(
	client *whatsmeow.Client,
	jid *types.JID,
	options *contract.Options,
	ctx *proto.ContextInfo,
) error {
	if options.Group.HiddenMention && utils.IsJidGroup(jid) {
		groupInfo, err := client.GetGroupInfo(*jid)
		if err != nil {
			return err
		}
		ctx.MentionedJid = make([]string, len(groupInfo.Participants))
		for i, v := range groupInfo.Participants {
			ctx.MentionedJid[i] = v.JID.String()
		}
	}

	return nil
}

func setQuotedMessage(ctx *proto.ContextInfo, quoted *contract.Quoted) {
	if quoted != nil && quoted.MessageId != "" {
		ctx.QuotedMessage = quoted.Message
		ctx.StanzaId = &quoted.MessageId
		ctx.Participant = &quoted.Sender
	}
}

func setMediaMessage(
	mediatype, caption, filename, mimetype string,
	contextInfo *proto.ContextInfo,
	upload *whatsmeow.UploadResponse,
	isGif bool, bytes []byte,
) (*proto.Message, error) {
	switch mediatype {
	case "image":
		return &proto.Message{
			ImageMessage: &proto.ImageMessage{
				Caption:             &caption,
				Mimetype:            &mimetype,
				ThumbnailDirectPath: &upload.DirectPath,
				ThumbnailSha256:     upload.FileSHA256,
				ThumbnailEncSha256:  upload.FileEncSHA256,
				FileLength:          &upload.FileLength,
				FileSha256:          upload.FileSHA256,
				FileEncSha256:       upload.FileEncSHA256,
				MediaKey:            upload.MediaKey,
				JpegThumbnail:       bytes,
				Url:                 &upload.URL,
				DirectPath:          &upload.DirectPath,
				ContextInfo:         contextInfo,
			},
		}, nil
	case "video":
		return &proto.Message{
			VideoMessage: &proto.VideoMessage{
				Caption:             &caption,
				Mimetype:            &mimetype,
				ThumbnailDirectPath: &upload.DirectPath,
				ThumbnailSha256:     upload.FileSHA256,
				ThumbnailEncSha256:  upload.FileEncSHA256,
				FileLength:          &upload.FileLength,
				FileSha256:          upload.FileSHA256,
				FileEncSha256:       upload.FileEncSHA256,
				MediaKey:            upload.MediaKey,
				GifPlayback:         &isGif,
				Url:                 &upload.URL,
				DirectPath:          &upload.DirectPath,
				ContextInfo:         contextInfo,
			},
		}, nil
	case "audio":
		ppt := true
		return &proto.Message{
			AudioMessage: &proto.AudioMessage{
				Mimetype:      &mimetype,
				FileLength:    &upload.FileLength,
				FileSha256:    upload.FileSHA256,
				FileEncSha256: upload.FileEncSHA256,
				MediaKey:      upload.MediaKey,
				Url:           &upload.URL,
				DirectPath:    &upload.DirectPath,
				Ptt:           &ppt,
				ContextInfo:   contextInfo,
			},
		}, nil
	case "document":
		if filename == "" {
			return nil, errors.New("the 'filename' is required for the 'document' type")
		}
		return &proto.Message{
			DocumentMessage: &proto.DocumentMessage{
				Caption:             &caption,
				Mimetype:            &mimetype,
				ThumbnailDirectPath: &upload.DirectPath,
				ThumbnailSha256:     upload.FileSHA256,
				ThumbnailEncSha256:  upload.FileEncSHA256,
				FileLength:          &upload.FileLength,
				FileSha256:          upload.FileSHA256,
				FileEncSha256:       upload.FileEncSHA256,
				MediaKey:            upload.MediaKey,
				Url:                 &upload.URL,
				DirectPath:          &upload.DirectPath,
				FileName:            &filename,
				ContextInfo:         contextInfo,
			},
		}, nil
	case "ptv":
		return &proto.Message{
			PtvMessage: &proto.VideoMessage{
				Caption:             &caption,
				Mimetype:            &mimetype,
				ThumbnailDirectPath: &upload.DirectPath,
				ThumbnailSha256:     upload.FileSHA256,
				ThumbnailEncSha256:  upload.FileEncSHA256,
				FileLength:          &upload.FileLength,
				FileSha256:          upload.FileSHA256,
				FileEncSha256:       upload.FileEncSHA256,
				MediaKey:            upload.MediaKey,
				GifPlayback:         &isGif,
				Url:                 &upload.URL,
				DirectPath:          &upload.DirectPath,
				ContextInfo:         contextInfo,
			},
		}, nil
	}

	return nil, errors.ErrUnsupported
}

func (s *Service) sendMessageWithTyping(
	instance *whatsapp.Instance,
	to *types.JID,
	message *proto.Message,
	options *contract.Options,
) (string, int, error) {

	client := instance.Client

	messageId := options.MessageID
	if messageId == "" {
		messageId = uuid.NewString()
	}

	go func() {
		if options.Delay != 0 {
			widthPresence := options.Presence != ""

			if widthPresence {
				var chatPresence types.ChatPresence
				var mediaPresence types.ChatPresenceMedia

				switch options.Presence {
				case "composing":
					chatPresence = types.ChatPresenceComposing
					mediaPresence = types.ChatPresenceMediaText
				case "recording":
					chatPresence = ""
					mediaPresence = types.ChatPresenceMediaAudio
				}

				err := client.SubscribePresence(*to)
				if err != nil {
					s.Logger.Error(err.Error())
				}

				err = client.SendChatPresence(*to, chatPresence, mediaPresence)
				if err != nil {
					s.Logger.Error(err.Error())
				}
			}

			time.Sleep(time.Duration(options.Delay) * time.Millisecond)

			if widthPresence {
				err := client.SendChatPresence(*to, types.ChatPresencePaused, types.ChatPresenceMediaText)
				if err != nil {
					s.Logger.Error(err.Error())
				}
			}
		}

		sent, err := client.SendMessage(s.Ctx, *to, message, whatsmeow.SendRequestExtra{
			ID: messageId,
		})

		if err != nil {
			fmt.Println("SEND ERROR: ", err)
			s.Messaging.SendMessage(string(messaging.SEND_MESSAGE), whatsapp.PreparedMessage(
				messaging.INSTANCE_ERROR,
				instance,
				map[string]any{
					"sent": map[string]any{
						"messageID": sent.ID,
						"timestamp": sent.Timestamp,
						"message":   message,
					},
					"Error": map[string]any{
						"IsError":     true,
						"Description": err.Error(),
					},
				},
			))
			return
		}

		s.Messaging.SendMessage(string(messaging.SEND_MESSAGE), whatsapp.PreparedMessage(
			messaging.SEND_MESSAGE,
			instance,
			map[string]any{
				"sent": map[string]any{
					"messageID": sent.ID,
					"timestamp": sent.Timestamp,
					"message":   message,
				},
				"Error": map[string]any{
					"IsError":     false,
				},
			},
		))
	}()

	return messageId, http.StatusCreated, nil
}

func (s *Service) TextMessage(param string, data *contract.TextMessage, quoted *contract.Quoted) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	var contextInfo proto.ContextInfo

	setQuotedMessage(&contextInfo, quoted)

	err = setMentionedGroup(instance.Client, &to, &data.Options, &contextInfo)
	if err != nil {
		return
	}

	message := proto.Message{
		ExtendedTextMessage: &proto.ExtendedTextMessage{
			Text:        &data.Message.Text,
			ContextInfo: &contextInfo,
		},
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, &message, &data.Options)
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}

	return
}

func (s *Service) LinkMessage(param string, data *contract.LinkMessage, quoted *contract.Quoted) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	var contextInfo proto.ContextInfo

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	setQuotedMessage(&contextInfo, quoted)
	err = setMentionedGroup(instance.Client, &to, &data.Options, &contextInfo)
	if err != nil {
		return
	}

	preview, err := utils.GeneratePreview(data.Message.Link)
	if err != nil {
		return
	}

	imageUrl := preview["image"]

	jpegThumbnail, _, err := utils.GetMediaByte(imageUrl)
	if err != nil {
		s.Logger.Error("Failed to generate jpegThumbnail:", err)
	}

	title := preview["title"]
	description := preview["description"]
	previewType := proto.ExtendedTextMessage_IMAGE

	text := data.Message.Link
	if data.Message.Text != "" {
		text = utils.StringJoin("\n\n", text, data.Message.Text)
	}

	message := proto.Message{
		ExtendedTextMessage: &proto.ExtendedTextMessage{
			Text:          &text,
			Title:         &title,
			Description:   &description,
			CanonicalUrl:  &data.Message.Link,
			MatchedText:   &data.Message.Link,
			PreviewType:   &previewType,
			JpegThumbnail: jpegThumbnail,
			ContextInfo:   &contextInfo,
		},
	}

	if data.Message.Title != "" {
		message.ExtendedTextMessage.Title = &data.Message.Title
	}
	if data.Message.Description != "" {
		message.ExtendedTextMessage.Description = &data.Message.Description
	}

	w, _ := strconv.ParseUint(preview["image:width"], 10, 32)
	h, _ := strconv.ParseUint(preview["image:height"], 10, 32)

	if w > 0 && h > 0 {
		width := uint32(w)
		height := uint32(h)

		message.ExtendedTextMessage.ThumbnailWidth = &width
		message.ExtendedTextMessage.ThumbnailHeight = &height
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, &message, &data.Options)
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}

	return
}

func (s *Service) MediaMessage(param, mime string, data *contract.MediaMessage, quoted *contract.Quoted) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	var contextInfo proto.ContextInfo

	bytes, headers, err := utils.GetMediaByte(data.Message.Url)
	if err != nil {
		return
	}

	upload, err := instance.Client.Upload(s.Ctx, bytes, utils.GetMediaType(data.Message.MediaType))
	if err != nil {
		return
	}

	if mime == "" {
		mime = headers.Get("Content-Type")
	}

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	setQuotedMessage(&contextInfo, quoted)
	err = setMentionedGroup(instance.Client, &to, &data.Options, &contextInfo)
	if err != nil {
		return
	}

	message, err := setMediaMessage(
		data.Message.MediaType, data.Message.Caption, data.Message.Filename,
		mime, &contextInfo, &upload, data.Message.GifPlayback, bytes,
	)
	if err != nil {
		return
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, message, &data.Options)
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}

	return
}

func (s *Service) FileMessage(param string, data *contract.MediaMessage, mimetype string, file []byte, quoted *contract.Quoted) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	var contextInfo proto.ContextInfo

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	setQuotedMessage(&contextInfo, quoted)

	err = setMentionedGroup(instance.Client, &to, &data.Options, &contextInfo)
	if err != nil {
		return
	}

	upload, err := instance.Client.Upload(s.Ctx, file, utils.GetMediaType(data.Message.MediaType))
	if err != nil {
		return
	}

	message, err := setMediaMessage(
		data.Message.MediaType, data.Message.Caption, data.Message.Filename,
		mimetype, &contextInfo, &upload, data.Message.GifPlayback, file,
	)
	if err != nil {
		return
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, message, &data.Options)
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}

	return
}

func (s *Service) LocationMessage(param string, data *contract.LocationMessage, quoted *contract.Quoted) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	var contextInfo proto.ContextInfo

	setQuotedMessage(&contextInfo, quoted)

	err = setMentionedGroup(instance.Client, &to, &data.Options, &contextInfo)
	if err != nil {
		return
	}

	// url := https://www.google.com/maps/search/?api=1&query=-20.3331946,-44.0604606
	url := utils.StringJoin(
		"",
		"https://www.google.com/maps/search/?api=1&query=",
		fmt.Sprintf("%f", data.Message.Latitude),
		",",
		fmt.Sprintf("%f", data.Message.Longitude),
	)
	message := proto.Message{
		LocationMessage: &proto.LocationMessage{
			DegreesLatitude:  &data.Message.Latitude,
			DegreesLongitude: &data.Message.Longitude,
			Name:             &data.Message.Name,
			Address:          &data.Message.Address,
			Url:              &url,
			Comment:          &data.Message.Comment,
			ContextInfo:      &contextInfo,
		},
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, &message, &data.Options)
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}
	return
}

func (s *Service) ContactMessage(param string, data *contract.ContactMessage, quoted *contract.Quoted) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	var contextInfo proto.ContextInfo
	var message proto.Message

	setQuotedMessage(&contextInfo, quoted)

	err = setMentionedGroup(instance.Client, &to, &data.Options, &contextInfo)
	if err != nil {
		return
	}

	vcard := func(contact *contract.Contact) string {
		return utils.StringJoin("",
			"BEGIN:VCARD\n",
			"VERSION:3.0\n",
			"FN:",
			contact.FullName,
			"\n",
			"item1.TEL;waid=",
			contact.RawNumber,
			":",
			contact.PhoneNumber,
			"\n",
			"item1.X-ABLabel:Celular\n",
			"END:VCARD",
		)
	}

	card := vcard(&data.Message[0])

	if len(data.Message) == 1 {
		message = proto.Message{
			ContactMessage: &proto.ContactMessage{
				DisplayName: &data.Message[0].FullName,
				Vcard:       &card,
			},
		}
	} else {
		displayName := utils.StringJoin(" ", fmt.Sprint(len(data.Message)), "contacts")
		if _, ok := strings.CutPrefix(data.Recipient, "55"); ok {
			displayName = utils.StringJoin(" ", fmt.Sprint(len(data.Message)), "contatos")
		}

		message = proto.Message{
			ContactsArrayMessage: &proto.ContactsArrayMessage{
				DisplayName: &displayName,
			},
		}

		for _, v := range data.Message {
			card := vcard(&v)
			message.ContactsArrayMessage.Contacts = append(message.ContactsArrayMessage.Contacts, &proto.ContactMessage{
				DisplayName: &v.FullName,
				Vcard:       &card,
			})
		}
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, &message, &data.Options)
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}

	return
}

// Deprecated: send list not work
//
// Error: server error status 405 - Method not allow
func (s *Service) ListMessage(param string, data *contract.ListMessage, quoted *contract.Quoted) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	to, _ := types.ParseJID((utils.FormatJid(data.Recipient)))

	var contextInfo proto.ContextInfo

	setQuotedMessage(&contextInfo, quoted)

	err = setMentionedGroup(instance.Client, &to, &data.Options, &contextInfo)
	if err != nil {
		return
	}

	listType := proto.ListMessage_SINGLE_SELECT
	sections := make([]*proto.ListMessage_Section, len(data.Message.Sections))
	for i, v := range data.Message.Sections {
		rows := make([]*proto.ListMessage_Row, len(v.Rows))
		for i, v := range v.Rows {
			rows[i] = &proto.ListMessage_Row{
				Title:       &v.Title,
				Description: &v.Description,
				RowId:       &v.RowId,
			}
		}

		sections[i] = &proto.ListMessage_Section{
			Title: &v.Title,
			Rows:  rows,
		}
	}

	message := proto.Message{
		ViewOnceMessageV2: &proto.FutureProofMessage{
			Message: &proto.Message{
				ListMessage: &proto.ListMessage{
					Title:       &data.Message.Title,
					Description: &data.Message.Description,
					ButtonText:  &data.Message.ButtonText,
					FooterText:  &data.Message.FooterText,
					ListType:    &listType,
					Sections:    sections,
				},
			},
		},
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, &message, &data.Options)
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}
	return
}

func (s *Service) PoolMessage(param string, data *contract.PollMessage, quoted *contract.Quoted) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	var contextInfo proto.ContextInfo

	setQuotedMessage(&contextInfo, quoted)

	err = setMentionedGroup(instance.Client, &to, &data.Options, &contextInfo)
	if err != nil {
		return
	}

	encKey := []byte("$apr1$x1fsl974$C8lw9YagOtjVyp5u8NzB/0")

	options := make([]*proto.PollCreationMessage_Option, len(data.Message.Options))
	for i, v := range data.Message.Options {
		options[i] = &proto.PollCreationMessage_Option{
			OptionName: &v,
		}
	}

	deviceListMetadataVersion := int32(2)
	message := proto.Message{
		PollCreationMessageV3: &proto.PollCreationMessage{
			EncKey:      encKey,
			Name:        &data.Message.Name,
			Options:     options,
			ContextInfo: &contextInfo,
		},
		MessageContextInfo: &proto.MessageContextInfo{
			DeviceListMetadataVersion: &deviceListMetadataVersion,
			MessageSecret:             encKey,
		},
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, &message, &data.Options)
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}
	return
}

func (s *Service) ReactionMessage(param string, data *contract.ReactionMessage) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	message := proto.Message{
		ReactionMessage: &proto.ReactionMessage{
			Key: &proto.MessageKey{
				Id:        &data.MessageKey.MessageID,
				RemoteJid: &data.MessageKey.ChatID,
				FromMe:    &data.MessageKey.IsFromMe,
			},
			Text: &data.Message.Reaction,
		},
	}

	id, status, err := s.sendMessageWithTyping(instance, &to, &message, nil)

	json = map[string]string{"messageId": id}

	return
}

func (s *Service) EditMessage(param, messageId string, data *contract.EditMessage) (json any, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.Manager.GetInstance(param)
	if err != nil {
		return
	}

	to, _ := types.ParseJID(utils.FormatJid(data.Recipient))

	newMessage := instance.Client.BuildEdit(to, messageId, &proto.Message{
		Conversation: &data.Message.NewContent,
	})

	id, status, err := s.sendMessageWithTyping(instance, &to, newMessage, &contract.Options{})
	if err != nil {
		return
	}

	json = map[string]string{"messageId": id}

	return
}
