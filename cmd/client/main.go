package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("There was an error connecting rabbit to client: %v", err)
	}
	defer conn.Close()
	fmt.Println("Connection was successful! Peril game client connected to RabbitMQ!")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("Could not create chanel: %v", err)
	}

	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Error getting username: %v", err)
	}
	gs := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+gs.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		handlerPause(gs))

	if err != nil {
		log.Fatalf("Could not subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+gs.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		handlerMove(gs, publishCh))

	if err != nil {
		log.Fatalf("Could not subscribe to army_moves: %v", err)
	}

	err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		routing.WarRecognitionsPrefix+".*",
		pubsub.SimpleQueueDurable,
		handlerWar(gs, publishCh))

	if err != nil {
		log.Fatalf("Could not subscribe to war: %v", err)
	}

	// REPL:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "move":
			mv, err := gs.CommandMove(words)
			if err != nil {
				fmt.Printf("Could not make move: %v\n", err)
				continue
			}
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+gs.GetUsername(),
				mv,
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
			}

			fmt.Println("Move published!")
		case "spawn":
			err = gs.CommandSpawn(words)
			if err != nil {
				fmt.Printf("Could not spawn: %v\n", err)
				continue
			}
		case "status":
			gs.CommandStatus()
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
