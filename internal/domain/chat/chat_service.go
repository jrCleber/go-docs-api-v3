package chat

import (
	"net/http"

	"codechat.dev/contract"
	"codechat.dev/internal/whatsapp"
	"codechat.dev/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Service struct {
	logger   *logrus.Entry
	instance *whatsapp.Instance
}

func NewService(instance *whatsapp.Instance) *Service {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &Service{
		logger:   logger.WithFields(logrus.Fields{"name": "chat-service"}),
		instance: instance,
	}
}

func (s *Service) ValidateWhatsAppNumbers(param string, number []string) (onWhatsapp []contract.IsOnWhatsAppResponse, status int, err error) {
	status = http.StatusBadRequest

	instance := s.instance

	jids := make([]string, len(number))
	for i := 0; i < len(number); i++ {
		jids[i] = utils.FormatJid(number[i])
	}

	resp, err := instance.Client.IsOnWhatsApp(jids)
	if err != nil {
		return
	}

	for _, v := range resp {
		onWhatsapp = append(onWhatsapp, contract.IsOnWhatsAppResponse{
			Query: v.Query,
			JID:   v.JID.String(),
			IsIn:  v.IsIn,
		})
	}

	return
}
