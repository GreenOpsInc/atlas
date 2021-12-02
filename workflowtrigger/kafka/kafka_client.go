package kafka

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"log"
	"strconv"
	"time"
)

type KafkaClient interface {
	SendMessage(data string) error
}

type KafkaClientImpl struct {
	kafkaWriter *kafka.Writer
}

const (
	kafkaTopic string = "greenops.eventing"
)

func New(address string) KafkaClient {
	k := KafkaClientImpl{}
	k.kafkaWriter = &kafka.Writer{
		Addr:     kafka.TCP(address),
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{},
	}
	return &k
}

func (k *KafkaClientImpl) SendMessage(data string) error {
	err := k.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Value: []byte(data),
		},
	)
	if err != nil {
		log.Printf("Failed to write messages: %s", err)
		return err
	}
	return nil
	//if err := k.kafkaWriter.Close(); err != nil {
	//	log.Fatal("failed to close writer:", err)
	//}
}

func produce(ctx context.Context) {
	// initialize a counter
	i := 0

	// intialize the writer with the broker addresses, and the topic
	w := &kafka.Writer{
		Addr:     kafka.TCP("localhost:9092"),
		Topic:    "topic-A",
		Balancer: &kafka.LeastBytes{},
	}

	for {
		// each kafka message has a key and value. The key is used
		// to decide which partition (and consequently, which broker)
		// the message gets published on
		err := w.WriteMessages(ctx, kafka.Message{
			Key: []byte(strconv.Itoa(i)),
			// create an arbitrary message payload for the value
			Value: []byte("this is message" + strconv.Itoa(i)),
		})
		if err != nil {
			panic("could not write message " + err.Error())
		}

		// log a confirmation once the message is written
		fmt.Println("writes:", i)
		i++
		// sleep for a second
		time.Sleep(time.Second)
	}
}
