package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	durable := simpleQueueType == SimpleQueueDurable
	autoDelete := simpleQueueType == SimpleQueueTransient
	exclusive := simpleQueueType == SimpleQueueTransient

	queue, err := channel.QueueDeclare(
		queueName,
		durable,
		autoDelete,
		exclusive,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(
		queue.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T),
) error {
	ch, q, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
	)
	if err != nil {
		return err
	}

	deliveries, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- consumeMessages(deliveries, handler)
	}()
	go func() {
		if err := <-errCh; err != nil {
			log.Printf("consumer terminated with error: %v", err)
		}
	}()

	return nil
}

func consumeMessages[T any](messages <-chan amqp.Delivery, handler func(T)) error {
	for message := range messages {
		var payload T
		if err := json.Unmarshal(message.Body, &payload); err != nil {
			if nackErr := message.Nack(false, true); nackErr != nil {
				log.Printf("failed to nack message: %v", nackErr)
			}
			continue
		}
		handler(payload)
		err := message.Ack(false)
		if err != nil {
			log.Printf("failed to ack message: %v", err)
			return err
		}
	}
	return nil
}
