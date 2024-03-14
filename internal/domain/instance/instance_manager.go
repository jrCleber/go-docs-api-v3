package instance

import (
	"fmt"

	"codechat.dev/internal/whatsapp"
	"codechat.dev/pkg/messaging"
	"github.com/sirupsen/logrus"
)

type MapInstances map[string]*whatsapp.Instance

type InstancesManager struct {
	logger    *logrus.Entry
	store     *whatsapp.Store
	Wa        MapInstances
	messaging *messaging.Amqp
}

func NewInstancesManager(store *whatsapp.Store, messaging *messaging.Amqp) *InstancesManager {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &InstancesManager{
		logger:    logger.WithFields(logrus.Fields{"name": "instances-manager"}),
		store:     store,
		Wa:        make(MapInstances),
		messaging: messaging,
	}
}

func (x *InstancesManager) GetInstance(param string) (*whatsapp.Instance, error) {
	find, err := x.store.Read(param)

	if err != nil {
		return nil, err
	}

	instance := x.Wa[find.ID]
	if instance == nil {
		return nil, fmt.Errorf("instance %s not found", param)
	}

	return instance, nil
}

func (x *InstancesManager) Load(service *Service) {
	instances, err := x.store.ReadAll()
	if err != nil {
		x.logger.Error(err)
		return
	}

	for _, instance := range instances {
		if instance.WhatsApp.Number == "" ||
			instance.State == whatsapp.INACTIVE {
			continue
		}
		_, connection, err := service.Connect(instance.ID)
		if err != nil {
			x.logger.Error(err)
			continue
		}

		x.Wa[connection.ID] = connection
		x.logger.WithFields(logrus.Fields{
			"instance": connection.Name,
			"number":   connection.WhatsApp.Number,
		}).Info("Connected")
	}
}
