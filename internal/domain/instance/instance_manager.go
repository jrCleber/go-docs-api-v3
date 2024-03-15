package instance

import (
	"codechat.dev/internal/whatsapp"
	"github.com/sirupsen/logrus"
)

type MapInstances map[string]*whatsapp.Instance

type InstancesManager struct {
	logger *logrus.Entry
}

func NewInstancesManager() *InstancesManager {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	return &InstancesManager{
		logger: logger.WithFields(logrus.Fields{"name": "instances-manager"}),
	}
}

func (x *InstancesManager) Load(service *Service) {
	_, connection, err := service.Connect()
	if err != nil {
		x.logger.Error(err)
	}

	x.logger.WithFields(logrus.Fields{
		"instance": connection.Name,
		"number":   connection.WhatsApp.Number,
	}).Info("Connected")
}
