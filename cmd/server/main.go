package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

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

	waitForShutdown()
	fmt.Println("Closing RabbitMQ connection...")
}

func waitForShutdown() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("\nReceived an interrupt, shutting down...")
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
