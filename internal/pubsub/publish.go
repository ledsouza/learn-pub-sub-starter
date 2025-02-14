package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"

	"github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp091.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp091.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
	if err != nil {
		return err
	}

	return nil
}

func PublishGob[T any](ch *amqp091.Channel, exchange, key string, val T) error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(val)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp091.Publishing{
		ContentType: "application/gob",
		Body:        buffer.Bytes(),
	})
	if err != nil {
		return err
	}

	return nil
}
