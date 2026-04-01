package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func handlerLogs() func(routing.GameLog) pubsub.Acktype {
	return func(gl routing.GameLog) pubsub.Acktype {
		defer fmt.Println("> ")

		err := gamelogic.WriteLog(gl)
		if err != nil {
			fmt.Printf("There was an error in writing the gamelog: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
