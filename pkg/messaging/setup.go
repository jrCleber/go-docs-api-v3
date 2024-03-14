package messaging

import (
	"encoding/json"

	"codechat.dev/pkg/config"
	"codechat.dev/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var ExchangeName = "codechat_api_v3"

type HandlerDelivery func(*amqp.Delivery)

type Amqp struct {
	Conn       *amqp.Connection
	cfg        *config.AMQP
	prefixName string
	logger     *logrus.Entry
}

func NewConnection(cfg *config.AMQP, containerName string) (*Amqp, error) {
	conn, err := amqp.Dial(cfg.Url)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()

	m := Amqp{
		Conn:       conn,
		cfg:        cfg,
		prefixName: containerName,
		logger:     logger.WithFields(logrus.Fields{"name": "messaging"}),
	}

	return &m, nil
}

func (m *Amqp) ConsumeMessages(queueName string, handler HandlerDelivery) {
	ch, err := m.Conn.Channel()
	if err != nil {
		m.logger.Error("Failed to open channel: ", err)
		return
	}

	msgs, err := ch.Consume(
		queueName,
		"",
		true,
		true,
		false,
		false,
		nil,
	)
	if err != nil {
		m.logger.Error("Failed to register a consumer: ", err)
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			handler(&d)
		}
	}()

	m.logger.Log(logrus.WarnLevel, "[*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func (m *Amqp) SetupExchangesAndQueues(events []string) {
	ch, err := m.Conn.Channel()
	if err != nil {
		m.logger.Error("Failed to open channel: ", err)
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil)

	if err != nil {
		m.logger.Error("Failed to declare an exchange: ", err)
		return
	}

	for _, q := range m.cfg.Queues {
		queueName := utils.StringJoin("_", q, m.prefixName)
		_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			m.logger.WithFields(logrus.Fields{
				"queue": queueName,
				"desc":  "queue declaration failed",
			}).Error(err)
			return
		}

		for _, evt := range events {
			err = ch.QueueBind(queueName, evt, ExchangeName, false, nil)
			if err != nil {
				m.logger.WithFields(logrus.Fields{
					"queue": queueName,
					"event": evt,
					"desc":  "failed to bind the queue",
				}).Error(err)
				return
			}

			m.logger.WithFields(logrus.Fields{
				"exchange": ExchangeName,
				"queue":    queueName,
				"event":    evt,
			}).Log(logrus.WarnLevel, "configuration complete")
		}
	}
}

func (m *Amqp) SendMessage(routingKey string, data any) {
	if data == nil {
		return
	}
	ch, err := m.Conn.Channel()
	if err != nil {
		m.logger.Error(err)
		return
	}
	defer ch.Close()

	msgBytes, err := json.Marshal(data)
	if err != nil {
		m.logger.Errorf("Failed to marshal struct: %v", err)
		return
	}

	err = ch.Publish(ExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        msgBytes,
	})

	if err != nil {
		m.logger.Errorf("Failed to publish a message: %v", err)
		return
	}

	m.logger.WithFields(logrus.Fields{
		"exchange":    ExchangeName,
		"routing-key": routingKey,
	}).Log(logrus.WarnLevel, "sent=true")
}

// func (m *Amqp) ExchangeDeleteAndQueues(exchangeName string) {
// 	ch, err := m.Conn.Channel()
// 	if err != nil {
// 		log.Println("error while connecting to the amqp server: ", err)
// 	}
// 	defer ch.Close()

// 	for _, q := range m.Cfg.Queues {
// 		_, err = ch.QueueDelete(utils.StringJoin("_", exchangeName, q), false, false, true)
// 		if err != nil {
// 			log.Println("queue deletion failed: ", err)
// 		}
// 	}

// 	err = ch.ExchangeDelete(exchangeName, false, true)
// 	if err != nil {
// 		log.Println("exchange deletion failed: ", err)
// 	}
// }
