package gamelogic

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	"github.com/rabbitmq/amqp091-go"
)

func PrintClientHelp() {
	fmt.Println("Possible commands:")
	fmt.Println("* move <location> <unitID> <unitID> <unitID>...")
	fmt.Println("    example:")
	fmt.Println("    move asia 1")
	fmt.Println("* spawn <location> <rank>")
	fmt.Println("    example:")
	fmt.Println("    spawn europe infantry")
	fmt.Println("* status")
	fmt.Println("* spam <n>")
	fmt.Println("    example:")
	fmt.Println("    spam 5")
	fmt.Println("* quit")
	fmt.Println("* help")
}

func ClientWelcome() (string, error) {
	fmt.Println("Welcome to the Peril client!")
	fmt.Println("Please enter your username:")
	words := GetInput()
	if len(words) == 0 {
		return "", errors.New("you must enter a username. goodbye")
	}
	username := words[0]
	fmt.Printf("Welcome, %s!\n", username)
	PrintClientHelp()
	return username, nil
}

func PrintServerHelp() {
	fmt.Println("Possible commands:")
	fmt.Println("* pause")
	fmt.Println("* resume")
	fmt.Println("* quit")
	fmt.Println("* help")
}

func GetInput() []string {
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	scanned := scanner.Scan()
	if !scanned {
		return nil
	}
	line := scanner.Text()
	line = strings.TrimSpace(line)
	return strings.Fields(line)
}

func GetMaliciousLog() string {
	possibleLogs := []string{
		"Never interrupt your enemy when he is making a mistake.",
		"The hardest thing of all for a soldier is to retreat.",
		"A soldier will fight long and hard for a bit of colored ribbon.",
		"It is well that war is so terrible, otherwise we should grow too fond of it.",
		"The art of war is simple enough. Find out where your enemy is. Get at him as soon as you can. Strike him as hard as you can, and keep moving on.",
		"All warfare is based on deception.",
	}
	randomIndex := rand.Intn(len(possibleLogs))
	msg := possibleLogs[randomIndex]
	return msg
}

func PrintQuit() {
	fmt.Println("I hate this game! (╯°□°)╯︵ ┻━┻")
}

func (gs *GameState) CommandStatus() {
	if gs.isPaused() {
		fmt.Println("The game is paused.")
		return
	} else {
		fmt.Println("The game is not paused.")
	}

	p := gs.GetPlayerSnap()
	fmt.Printf("You are %s, and you have %d units.\n", p.Username, len(p.Units))
	for _, unit := range p.Units {
		fmt.Printf("* %v: %v, %v\n", unit.ID, unit.Location, unit.Rank)
	}
}

func (gs *GameState) CommandSpam(input []string, channel *amqp091.Channel) {
	if len(input) < 2 {
		fmt.Println("spam <n>")
		return
	}

	count, err := strconv.Atoi(input[1])
	if err != nil {
		fmt.Println("Invalid number provided")
		return
	}

	for i := 0; i < count; i++ {
		maliciousLog := GetMaliciousLog()
		pubsub.PublishGob(
			channel,
			routing.ExchangePerilTopic,
			routing.GameLogSlug+"."+gs.GetPlayerSnap().Username,
			routing.GameLog{
				CurrentTime: time.Now(),
				Message:     maliciousLog,
				Username:    gs.GetPlayerSnap().Username,
			},
		)
	}
}
