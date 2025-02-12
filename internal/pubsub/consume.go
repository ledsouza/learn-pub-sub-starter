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
	handler func(T) AckType,
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

func consumeMessages[T any](messages <-chan amqp.Delivery, handler func(T) AckType) error {
	for message := range messages {
		var payload T
		if err := json.Unmarshal(message.Body, &payload); err != nil {
			if nackErr := message.Nack(false, true); nackErr != nil {
				log.Printf("failed to nack message: %v", nackErr)
			}
			continue
		}
		acktype := handler(payload)
		switch acktype {
		case Ack:
			log.Printf("Processing acktype %v\n", acktype)
			err := message.Ack(false)
			return handleAckTypeError(acktype, err)
		case NackRequeue:
			log.Printf("Processing acktype %v\n", acktype)
			err := message.Nack(false, true)
			return handleAckTypeError(acktype, err)
		case NackDiscard:
			log.Printf("Processing acktype %v\n", acktype)
			err := message.Nack(false, false)
			return handleAckTypeError(acktype, err)
		default:
			log.Printf("unknown AckType: %v; discarding message", acktype)
			err := message.Nack(false, false)
			return handleAckTypeError(NackDiscard, err)
		}
	}
	return nil
}

func handleAckTypeError(ackType AckType, err error) error {
	if err != nil {
		log.Printf("failed to process AckType %v: %v", ackType, err)
		return err
	}
	return nil
}
