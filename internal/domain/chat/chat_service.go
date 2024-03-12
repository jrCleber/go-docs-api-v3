package chat

import (
	"net/http"

	"codechat.dev/contract"
	"codechat.dev/internal/domain/instance"
	"codechat.dev/internal/whatsapp"
	"codechat.dev/pkg/messaging"
	"codechat.dev/pkg/utils"
	"github.com/sirupsen/logrus"
)

type Service struct {
	logger    *logrus.Entry
	store     *whatsapp.Store
	manager   *instance.InstancesManager
	messaging *messaging.Amqp
}

func NewService(store *whatsapp.Store, manager *instance.InstancesManager, messaging *messaging.Amqp) *Service {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &Service{
		logger:    logger.WithFields(logrus.Fields{"name": "chat-service"}),
		store:     store,
		manager:   manager,
		messaging: messaging,
	}
}

func (s *Service) ValidateWhatsAppNumbers(param string, number []string) (onWhatsapp []contract.IsOnWhatsAppResponse, status int, err error) {
	status = http.StatusBadRequest

	instance, err := s.manager.GetInstance(param)
	if err != nil {
		return
	}

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
			JID: v.JID.String(),
			IsIn: v.IsIn,
		})
	}

	return
}
