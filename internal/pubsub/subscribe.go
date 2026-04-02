package pubsub

import (
	"bytes"
	"encoding/gob"
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
	amqpCh, amqpQueue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	err = amqpCh.Qos(10, 0, false)
	if err != nil {
		fmt.Printf("Error in setting prefetch count: %v", err)
	}

	deliveries, err := amqpCh.Consume(amqpQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		defer amqpCh.Close()
		for delivery := range deliveries {
			var v T
			err := json.Unmarshal(delivery.Body, &v)
			if err != nil {
				fmt.Printf("We encountered an error: %v", err)
				continue
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

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) Acktype,
) error {
	amqpCh, amqpQueue, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}

	err = amqpCh.Qos(10, 0, false)
	if err != nil {
		fmt.Printf("Error in setting prefetch count: %v", err)
	}

	deliveries, err := amqpCh.Consume(amqpQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		defer amqpCh.Close()
		for delivery := range deliveries {
			b := bytes.NewBuffer(delivery.Body)
			dec := gob.NewDecoder(b)
			var v T
			err := dec.Decode(&v)
			if err != nil {
				fmt.Printf("We encountered an error: %v", err)
				continue
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
