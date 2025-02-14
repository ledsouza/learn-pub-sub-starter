package main

import (
	"fmt"
	"log"

	gamelogic "github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	pubsub "github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	routing "github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	const rabbitConnString = "amqp://guest:guest@localhost:5672/"

	fmt.Println("Starting Peril server...")
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
	}
	defer conn.Close()
	fmt.Println("Successfully connected to RabbitMQ")

	channel, err := conn.Channel()
	if err != nil {
		failOnError(err, "Failed to create channel")
	}

	pubsub.PublishJSON(
		channel,
		routing.ExchangePerilDirect,
		routing.PauseKey,
		routing.PlayingState{IsPaused: true},
	)

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		handlerLog(),
	)
	if err != nil {
		failOnError(err, "Failed to subscribe game_logs queue")
	}

	gamelogic.PrintServerHelp()

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		firstWord := input[0]
		switch firstWord {
		case "pause":
			fmt.Println("Sending pause message...")
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: true},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "resume":
			fmt.Println("Sending resume message...")
			err := pubsub.PublishJSON(
				channel,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{IsPaused: false},
			)
			if err != nil {
				log.Printf("could not publish time: %v", err)
			}
		case "quit":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func handlerLog() func(routing.GameLog) pubsub.AckType {
	return func(gl routing.GameLog) pubsub.AckType {
		defer fmt.Print("> ")
		err := gamelogic.WriteLog(gl)
		if err != nil {
			return pubsub.NackDiscard
		}
		return pubsub.Ack
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
