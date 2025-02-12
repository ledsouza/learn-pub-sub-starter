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

	fmt.Println("Starting Peril client...")

	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		failOnError(err, "Failed to connect to RabbitMQ")
	}
	defer conn.Close()

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		failOnError(err, "Failed to get username")
	}
	pauseQueueName := fmt.Sprintf("%s.%s", routing.PauseKey, username)
	movesQueueName := fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, username)

	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		pauseQueueName,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		failOnError(err, "Failed to declare and bind queue")
	}

	movesCh, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilTopic,
		movesQueueName,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		failOnError(err, "Failed to declare and bind queue")
	}

	gameState := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		pauseQueueName,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gameState),
	)
	if err != nil {
		failOnError(err, "Failed to subscribe to queue")
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		movesQueueName,
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gameState),
	)
	if err != nil {
		failOnError(err, "Failed to subscribe to queue")
	}

	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		cmd := input[0]
		switch cmd {
		case "spawn":
			err := gameState.CommandSpawn(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "move":
			armyMove, err := gameState.CommandMove(input)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(
				movesCh,
				routing.ExchangePerilTopic,
				movesQueueName,
				armyMove,
			)
			if err != nil {
				fmt.Println(err)
				continue
			}
		case "status":
			gameState.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Println("Unknown command")
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		return gs.HandlePause(ps)
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(gl gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		_, acktype := gs.HandleMove(gl)
		return acktype
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
