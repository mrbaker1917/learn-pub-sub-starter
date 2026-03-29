package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	amqpCh, _, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	deliveries, err := amqpCh.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for delivery := range deliveries {
			var v T
			err := json.Unmarshal(delivery.Body, &v)
			if err != nil {
				fmt.Printf("We encountered an error: %v", err)
			}
			switch handler(v) {
			case Ack:
				delivery.Ack(false)
				fmt.Println("Ack")
			case NackRequeue:
				delivery.Nack(false, true)
				fmt.Println("NackRequeue")
			case NackDiscard:
				delivery.Nack(false, false)
				fmt.Println("NackDiscard")
			}
		}
	}()

	return nil
}
