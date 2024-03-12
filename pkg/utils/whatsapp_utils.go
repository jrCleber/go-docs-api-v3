package utils

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
)


func ComparingTypes(key, value string) bool {
	return key == value
}

func GetContentType(message interface{}) string {
	val := reflect.ValueOf(message).Elem()

	var t string

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := val.Type().Field(i)

		if field.Kind() == reflect.Ptr && !field.IsNil() {
			t = typeField.Name
			break
		}
	}

	t = strings.Replace(t, "Message", "", 1)

	return t
}

func GetMap(data interface{}) map[string]interface{} {
	var decode map[string]any
	mapstructure.Decode(data, &decode)
	return decode
}

func GetSimpleMessage(m *proto.Message) (string, interface{}) {

	contentType := GetContentType(m)
	content := GetMap(m)[contentType]

	switch content := content.(type) {
	case *string:
		return contentType, *content
	case *proto.Message:
		return contentType, content
	default:
		return contentType, content
	}
}

func GetMessageTypeAndContent(msg *proto.Message) (string, interface{}) {
	contentType := GetContentType(msg)
	content := GetMap(msg)

	switch contentType {
	case "Ephemeral", "ViewOnce", "ViewOnceMessageV2":
		if m, ok := content[contentType].(*proto.FutureProofMessage); ok {
			return GetSimpleMessage(m.Message)
		}
		return "", nil
	case "Edited":
		_, content := GetSimpleMessage(msg.EditedMessage.Message)
		return contentType, content
	case "Conversation":
		contentType = "text"
		content = map[string]interface{}{"text": content["Conversation"]}
		return contentType, content
	case "ExtendedText":
		if *msg.ExtendedTextMessage.CanonicalUrl != "" || *msg.ExtendedTextMessage.MatchedText != "" {
			contentType = "link"
		} else {
			contentType = "text"
		}
		return contentType, content["ExtendedTextMessage"]
	default:
		return GetSimpleMessage(msg)
	}
}

func AdjustingJid(jid types.JID) string {
	input := jid.String()

	start := strings.Index(input, ":")
	end := strings.Index(input, "@")

	if start == -1 || end == -1 || end < start {
		return input
	}

	return input[:start] + input[end:]
}

func FormattedBrNumber(n string) string {
	regexp := regexp.MustCompile(`^(\d{2})(\d{2})\d{1}(\d{8})$`)
	match := regexp.FindStringSubmatch(n)

	if match != nil {
		if match[1] == "55" {
			joker, _ := strconv.Atoi(string(match[3][0]))
			ddd, _ := strconv.Atoi(match[2])
			if joker < 7 || ddd < 31 {
				return match[0]
			}
			return StringJoin("", match[1], match[2], match[3])
		}
	}

	return n
}

func FormatJid(n string) string {
	if strings.Contains(n, "@us") || strings.Contains(n, "@s.whatsapp.net") {
		return n
	}

	if strings.Contains(n, "-") {
		return StringJoin("", n, "@g,us")
	}

	brNumber := FormattedBrNumber(n)
	if brNumber != n {
		return StringJoin("", brNumber, "@s.whatsapp.net")
	}

	return StringJoin("", n, "@s.whatsapp.net")
}

func IsJidGroup(jid *types.JID) bool {
	return strings.Contains(jid.String(), "@g.us")
}

func GetMediaType(t string) whatsmeow.MediaType {
	switch t {
	case "image":
		return whatsmeow.MediaImage
	case "video":
		return whatsmeow.MediaVideo
	case "audio":
		return whatsmeow.MediaAudio
	case "document":
		return whatsmeow.MediaDocument
	case "sticker":
		return whatsmeow.MediaImage
	case "ptv":
		return whatsmeow.MediaVideo
	}

	return ""
}
