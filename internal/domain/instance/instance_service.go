package instance

import (
	"fmt"
	"net/http"
	"time"

	"codechat.dev/internal/whatsapp"
	"codechat.dev/pkg/messaging"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

type Service struct {
	logger   *logrus.Entry
	instance *whatsapp.Instance
	msRoute  string
}

func NewService(instance *whatsapp.Instance, messaging *messaging.Amqp, msRoute string) *Service {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &Service{
		logger:   logger.WithFields(logrus.Fields{"name": "instance-service"}),
		instance: instance,
		msRoute:  msRoute,
	}
}

func (s *Service) Find() (instance *whatsapp.Instance, statusCode int, err error) {
	instance = s.instance

	instance.Apikey = nil

	statusCode = http.StatusOK

	return
}

func (s *Service) NewConnection() (qrcode *whatsapp.QrCode, statusCode int, instance *whatsapp.Instance, err error) {
	statusCode = http.StatusBadRequest

	instance = s.instance

	err = instance.NewConnection()
	if err != nil {
		s.logger.Errorf("Unable to connect instance %s:\n Error: %v", instance.Name, err)
		return
	}

	time.Sleep(2 * time.Second)

	qrcode = instance.QrCode()

	return
}

func (s *Service) Connect() (statusCode int, instance *whatsapp.Instance, err error) {
	statusCode = http.StatusBadRequest
	instance = s.instance

	err = instance.Connect()
	if err != nil {
		s.logger.Errorf("Unable to connect instance %s:\n Error: %v", instance.Name, err)
		return
	}

	return
}

func (s *Service) FindDevice() {}

func (s *Service) Logout() (instance *whatsapp.Instance, statusCode int, err error) {
	instance = s.instance

	err = instance.Client.Logout()
	if err != nil {
		statusCode = http.StatusInternalServerError
		instance.Client.Disconnect()
		return
	}

	update := time.Now()
	instance.Connection = whatsapp.CLOSE
	instance.Status = whatsapp.Waiting
	instance.UpdateAt = &update

	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	return
}

func (s *Service) Delete(query string) (instance *whatsapp.Instance, statusCode int, err error) {
	statusCode = http.StatusBadRequest

	instance = s.instance

	if instance.Client != nil && (instance.Client.IsConnected() || instance.Client.IsLoggedIn()) {
		err = fmt.Errorf(
			"instance '%s' {jid: '%s', id: '%s'} is connected",
			instance.Name, instance.WhatsApp.Number, instance.ID,
		)
		return
	}

	update := time.Now()
	instance.DeletedAt = &update
	statusCode = http.StatusOK

	go func() {
		s.instance.Messaging.SendMessage(
			string(messaging.INSTANCE_STATUS),
			whatsapp.PreparedMessage(messaging.INSTANCE_STATUS, instance, map[string]any{
				"Status":    whatsapp.Deleted,
				"DeletedAt": time.Now(),
			}))
	}()

	return
}

func (s *Service) UpdateProfileName(query, value string) (statusCode int, err error) {
	statusCode = http.StatusBadRequest

	instance := s.instance
	err = instance.Client.SetGroupName(types.JID{}, value)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusNoContent

	return
}
