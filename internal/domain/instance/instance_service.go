package instance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"codechat.dev/contract"
	"codechat.dev/internal/whatsapp"
	"codechat.dev/pkg/messaging"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
)

type Service struct {
	logger    *logrus.Entry
	store     *whatsapp.Store
	manager   *InstancesManager
	messaging *messaging.Amqp
	msRoute   string
}

func NewService(store *whatsapp.Store, manager *InstancesManager, messaging *messaging.Amqp, msRoute string) *Service {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	return &Service{
		logger:    logger.WithFields(logrus.Fields{"name": "instance-service"}),
		store:     store,
		manager:   manager,
		messaging: messaging,
		msRoute:   msRoute,
	}
}

func (s *Service) Create(props *contract.Instance) (instance *whatsapp.Instance, statusCode int, err error) {
	// if microservice
	statusCode = http.StatusBadRequest

	find, _ := s.store.Read(props.Name)

	if find != nil || strings.ToLower(props.Name) == "codechat" {
		statusCode = http.StatusConflict
		err = fmt.Errorf("it is not possible to create an instance with this name: %s", props.Name)
		return
	}

	var apikey, containerName string
	if props.ApiKey == nil {
		apikey = ""
	}

	instance = whatsapp.NewInstance(
		"",
		props.Name,
		props.ExternalId,
		"",
		apikey,
		whatsapp.StateEnum(props.State),
		s.messaging,
		s.store,
		containerName,
	)

	body, err := json.Marshal(instance)
	if err != nil {
		return
	}

	res, err := http.Post(s.msRoute, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return
	}

	if res.StatusCode != http.StatusCreated {
		statusCode = res.StatusCode
		var body map[string]any
		err = render.DecodeJSON(res.Body, &body)
		if err != nil {
			return
		}
		if body != nil {
			message := body["message"].([]any)[0].(string)
			err = errors.New(message)
			return
		}
	}

	err = s.store.Create(
		instance.ID, instance.Name,
		"",
		*instance.Apikey,
		string(instance.State),
		string(instance.Connection),
		containerName,
		instance.CreatedAt,
	)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}
	statusCode = http.StatusCreated

	go func() {
		s.messaging.SendMessage(
			string(messaging.INSTANCE_STATUS),
			whatsapp.PreparedMessage(messaging.INSTANCE_STATUS, instance, map[string]whatsapp.StatusEnum{
				"Status": whatsapp.Created,
			}))
	}()

	return
}

func (s *Service) Find(query string) (instance *whatsapp.Instance, statusCode int, err error) {
	// if microservice

	instance, err = s.store.Read(query)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	if instance == nil {
		statusCode = http.StatusBadRequest
		err = fmt.Errorf("instance %s not found", query)
		return
	}

	if s.manager.Wa != nil {
		data, ok := s.manager.Wa[instance.ID]
		if ok {
			instance = data.GetInstance()
		}
	}

	instance.Apikey = nil

	statusCode = http.StatusOK

	return
}

func (s *Service) FindAll() (instances []*whatsapp.Instance, statusCode int, err error) {
	instances, err = s.store.ReadAll()
	if err != nil {
		return []*whatsapp.Instance{}, http.StatusOK, nil
	}

	return instances, http.StatusOK, nil
}

func (s *Service) NewConnection(param string) (qrcode *whatsapp.QrCode, statusCode int, instance *whatsapp.Instance, err error) {
	// if microservice

	statusCode = http.StatusBadRequest

	instance, statusCode, err = s.Find(param)
	if err != nil {
		return
	}

	instance.Messaging = s.messaging
	instance.Store = s.store

	err = instance.NewConnection()
	if err != nil {
		s.logger.Errorf("Unable to connect instance %s:\n Error: %v", instance.Name, err)
		return
	}

	s.manager.Wa[instance.ID] = instance

	time.Sleep(2 * time.Second)

	qrcode = instance.QrCode()

	return
}

func (s *Service) Connect(param string) (statusCode int, instance *whatsapp.Instance, err error) {
	// if microservice

	statusCode = http.StatusBadRequest

	instance, statusCode, err = s.Find(param)
	if err != nil {
		return
	}

	instance.Messaging = s.messaging
	instance.Store = s.store

	err = instance.Connect()
	if err != nil {
		s.logger.Errorf("Unable to connect instance %s:\n Error: %v", instance.Name, err)
		return
	}

	s.manager.Wa[instance.ID] = instance

	return
}

func (s *Service) Logout(query string) (instance *whatsapp.Instance, statusCode int, err error) {
	// if microservice

	find, statusCode, err := s.Find(query)
	if err != nil {
		return
	}

	instance = s.manager.Wa[find.ID]

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

	err = s.store.Update(find.ID, &whatsapp.Instance{
		UpdateAt:   instance.UpdateAt,
		Connection: instance.Connection,
		Status:     instance.Status,
	})
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	return
}

func (s *Service) Delete(query string) (instance *whatsapp.Instance, statusCode int, err error) {
	// if microservice
	statusCode = http.StatusBadRequest

	find, err := s.store.Read(query)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	if find == nil {
		err = fmt.Errorf("instance %s not found", query)
		return
	}

	if instance = s.manager.Wa[find.ID]; instance != nil {
		if instance.Client.IsConnected() || instance.Client.IsLoggedIn() {
			err = fmt.Errorf(
				"instance '%s' {jid: '%s', id: '%s'} is connected",
				find.Name, find.WhatsApp.Number, find.ID,
			)
			return
		}
	} else {
		instance = find
	}

	err = s.store.Delete(find.ID)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	delete(s.manager.Wa, find.ID)

	update := time.Now()
	instance.DeletedAt = &update
	statusCode = http.StatusOK

	go func() {
		s.messaging.SendMessage(
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

	find, _, err := s.Find(query)
	if err != nil {
		return
	}

	instance := s.manager.Wa[find.ID]
	err = instance.Client.SetGroupName(types.JID{}, value)
	if err != nil {
		statusCode = http.StatusInternalServerError
		return
	}

	statusCode = http.StatusNoContent

	return
}

func (s *Service) UpdateWhatsAppNumber(query, value string) (statusCode int, err error) {
	statusCode = http.StatusBadRequest

	find, _, err := s.Find(query)
	if err != nil {
		return
	}

	err = s.store.Update(find.ID, &whatsapp.Instance{WhatsApp: &whatsapp.WhatsApp{Number: value}})
	if err != nil {
		return
	}

	statusCode = http.StatusNoContent

	return
}
