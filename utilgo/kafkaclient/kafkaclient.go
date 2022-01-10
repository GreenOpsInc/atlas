package kafkaclient

import (
	"context"
	"crypto/tls"
	"log"

	"github.com/greenopsinc/util/tlsmanager"

	"github.com/segmentio/kafka-go"
)

type KafkaClient interface {
	SendMessage(data string) error
}

type kafkaClient struct {
	address     string
	kafkaWriter *kafka.Writer
	tm          tlsmanager.Manager
}

const (
	kafkaTopic string = "greenops.eventing"
)

func New(address string, tm tlsmanager.Manager) (KafkaClient, error) {
	k := &kafkaClient{address: address, tm: tm}
	writer, err := k.initWriter()
	if err != nil {
		return nil, err
	}
	k.kafkaWriter = writer
	return k, nil
}

func (k *kafkaClient) SendMessage(data string) error {
	err := k.kafkaWriter.WriteMessages(context.Background(),
		kafka.Message{
			Value: []byte(data),
		},
	)
	if err != nil {
		log.Printf("Failed to write messages: %s", err)
		return err
	}
	return k.kafkaWriter.Close()
}

func (k *kafkaClient) initWriter() (*kafka.Writer, error) {
	log.Println("in kafka initWriter")
	tlsConf, err := k.tm.GetClientTLSConf(tlsmanager.ClientKafka)
	log.Println("received kafka tls conf ", tlsConf)
	if err != nil {
		return nil, err
	}
	writer := k.configureWriter(tlsConf)
	log.Println("configured kafka writer ", writer)
	if err = k.watchWriter(); err != nil {
		return nil, err
	}
	log.Println("started kafka watcher err = ", err)
	return writer, nil
}

func (k *kafkaClient) configureWriter(tlsConf *tls.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(k.address),
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			TLS: tlsConf,
		},
	}
}

func (k *kafkaClient) watchWriter() error {
	err := k.tm.WatchClientTLSConf(tlsmanager.ClientKafka, func(conf *tls.Config, err error) {
		log.Printf("in watchClient, conf = %v, err = %v\n", conf, err)
		if err != nil {
			log.Fatalf("an error occurred in the watch %s client: %s", tlsmanager.ClientKafka, err.Error())
		}
		k.kafkaWriter = k.configureWriter(conf)
	})
	return err
}
