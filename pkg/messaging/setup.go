package messaging

import (
	"encoding/json"
	"log"

	"codechat.dev/pkg/config"
	"codechat.dev/pkg/utils"
	"github.com/streadway/amqp"
)

var ExchangeName = "codechat_api_v3"

type HandlerDelivery func(*amqp.Delivery)

type Amqp struct {
	Conn *amqp.Connection
	Cfg  *config.AMQP
}

func NewConnection(cfg *config.AMQP) (*Amqp, error) {
	conn, err := amqp.Dial(cfg.Url)
	if err != nil {
		return nil, err
	}

	m := Amqp{Conn: conn, Cfg: cfg}

	return &m, nil
}

func (m *Amqp) ConsumeMessages(queueName string, handler HandlerDelivery) {
	ch, err := m.Conn.Channel()
	if err != nil {
		log.Println("Failed to open channel: ", err)
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
		log.Println("Failed to register a consumer: ", err)
		return
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			handler(&d)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func (m *Amqp) SetupExchangesAndQueues(events []string) {
	ch, err := m.Conn.Channel()
	if err != nil {
		log.Println("Failed to open channel: ", err)
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(ExchangeName, "topic", true, false, false, false, nil)

	if err != nil {
		log.Println("Failed to declare an exchange: ", err)
		return
	}

	for _, q := range m.Cfg.Queues {
		queueName := utils.StringJoin("_", ExchangeName, q)
		_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
		if err != nil {
			log.Printf("Failed to declare a queue for %s: %v", queueName, err)
			return
		}

		for _, evt := range events {
			err = ch.QueueBind(queueName, evt, ExchangeName, false, nil)
			if err != nil {
				log.Printf("Failed to bind a queue for %s: %v", evt, err)
				return
			}

			log.Printf("Setup complete for exchange name: %s - Evt: %s", ExchangeName, evt)
		}
	}

	return
}

func (m *Amqp) SendMessage(routingKey string, data any) {
	if data == nil {
		return
	}
	ch, err := m.Conn.Channel()
	if err != nil {
		log.Println("Failed to open channel: ", err)
		return
	}
	defer ch.Close()

	msgBytes, err := json.Marshal(data)
	if err != nil {
		log.Printf("Failed to marshal struct: %v", err)
		return
	}

	err = ch.Publish(ExchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        msgBytes,
	})

	if err != nil {
		log.Printf("Failed to publish a message: %v", err)
		return
	}

	log.Printf("Message sent to exchange %s with routing key %s", ExchangeName, routingKey)
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
