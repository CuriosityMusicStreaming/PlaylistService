package integrationevent

import (
	"github.com/CuriosityMusicStreaming/ComponentsPool/pkg/app/storedevent"
	commonamqp "github.com/CuriosityMusicStreaming/ComponentsPool/pkg/infrastructure/amqp"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

const (
	domainEventExchangeName = "domain_event"
	domainEventExchangeType = "topic"
	domainEventsQueueName   = "playlist_service_domain_event"

	contentType = "application/json; charset=utf-8"

	transportName = "amqp_integration_events"
)

type Handler interface {
	Handle(msgBody string) error
}

type Transport interface {
	commonamqp.Channel
	storedevent.Transport
	SetHandler(handler Handler)
}

func NewIntegrationEventTransport() Transport {
	return &transport{}
}

type transport struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	handler Handler
}

func (t *transport) SetHandler(handler Handler) {
	t.handler = handler
}

func (t *transport) Name() string {
	return transportName
}

func (t *transport) Send(_, msgBody string) error {
	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  contentType,
		Body:         []byte(msgBody),
	}
	return t.channel.Publish(domainEventExchangeName, "", false, false, msg)
}

func (t *transport) Connect(conn *amqp.Connection) error {
	t.conn = conn

	channel, err := conn.Channel()
	if err != nil {
		return err
	}

	t.channel = channel

	err = channel.ExchangeDeclare(domainEventExchangeName, domainEventExchangeType, true, false, false, false, nil)
	if err != nil {
		return err
	}

	if t.handler == nil {
		return errors.New("cannot connect to read channel without handler")
	}

	return t.connectToReadChannel()
}

func (t *transport) connectToReadChannel() error {
	queue, err := t.channel.QueueDeclare(domainEventsQueueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = t.channel.QueueBind(queue.Name, "", domainEventExchangeName, false, nil)
	if err != nil {
		return err
	}

	readChan, err := t.channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range readChan {
			err = t.handler.Handle(string(delivery.Body))
			if err == nil {
				err = delivery.Ack(false)
			} else {
				err = delivery.Nack(false, true)
			}
			_ = err
		}
	}()

	return nil
}
