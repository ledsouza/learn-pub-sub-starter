package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
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
		amqp.Table{
			"x-dead-letter-exchange": "peril_dlx",
		},
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
	err := subscribe(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		consumeJSONMessages,
	)
	if err != nil {
		return err
	}
	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
) error {
	err := subscribe(
		conn,
		exchange,
		queueName,
		key,
		simpleQueueType,
		handler,
		consumeGobMessages,
	)
	if err != nil {
		return err
	}
	return nil
}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	messageConsumer func(<-chan amqp.Delivery, func(T) AckType) error,
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

	err = ch.Qos(10, 0, true)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		defer ch.Close()
		errCh <- messageConsumer(msgs, handler)
	}()
	go func() {
		if err := <-errCh; err != nil {
			log.Fatalf("consumer terminated with fatal error: %v", err)
		}
	}()

	return nil
}

func consumeJSONMessages[T any](messages <-chan amqp.Delivery, handler func(T) AckType) error {
	for message := range messages {
		var payload T
		if err := json.Unmarshal(message.Body, &payload); err != nil {
			if nackErr := message.Nack(false, true); nackErr != nil {
				log.Printf("failed to nack message: %v", nackErr)
			}
			continue
		}
		acktype := handler(payload)
		err := handleMessageAck(message, acktype)
		if err != nil {
			return err
		}
	}
	return nil
}

func consumeGobMessages[T any](messages <-chan amqp.Delivery, handler func(T) AckType) error {
	for message := range messages {
		buffer := bytes.NewBuffer(message.Body)
		var payload T
		decoder := gob.NewDecoder(buffer)
		err := decoder.Decode(&payload)
		if err != nil {
			if nackErr := message.Nack(false, true); nackErr != nil {
				log.Printf("failed to nack message: %v", nackErr)
			}
			continue
		}
		acktype := handler(payload)
		err = handleMessageAck(message, acktype)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleMessageAck(message amqp.Delivery, acktype AckType) error {
	switch acktype {
	case Ack:
		err := message.Ack(false)
		if err != nil {
			return handleAckTypeError(acktype, err)
		}
	case NackRequeue:
		err := message.Nack(false, true)
		if err != nil {
			return handleAckTypeError(acktype, err)
		}
	case NackDiscard:
		err := message.Nack(false, false)
		if err != nil {
			return handleAckTypeError(acktype, err)
		}
	default:
		log.Printf("unknown AckType: %v; discarding message", acktype)
		err := message.Nack(false, false)
		if err != nil {
			return handleAckTypeError(acktype, err)
		}
	}
	return nil
}

func handleAckTypeError(ackType AckType, err error) error {
	log.Printf("failed to process AckType %v: %v", ackType, err)
	return err
}
