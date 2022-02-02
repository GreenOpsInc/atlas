package kafkaclient

import (
	"context"
	"crypto/tls"
	"log"
	"os"

	"github.com/greenopsinc/util/tlsmanager"

	"github.com/segmentio/kafka-go"
)

type KafkaClient interface {
	SendMessage(data string) error
}

type kafkaClient struct {
	address string
	tlsConf *tls.Config
	tm      tlsmanager.Manager
}

const (
	defaultKafkaTopic string = "greenops.eventing"
)

func New(address string, tm tlsmanager.Manager) (KafkaClient, error) {
	k := &kafkaClient{address: address, tm: tm}
	if err := k.initWriter(); err != nil {
		return nil, err
	}
	return k, nil
}

func (k *kafkaClient) SendMessage(data string) error {
	writer, err := k.configureWriter(k.tlsConf)
	if err != nil {
		return err
	}
	err = writer.WriteMessages(context.Background(),
		kafka.Message{
			Value: []byte(data),
		},
	)
	if err != nil {
		log.Printf("Failed to write messages: %s", err)
		return err
	}
	return writer.Close()
}

func (k *kafkaClient) initWriter() error {
	tlsConf, err := k.tm.GetKafkaTLSConf()
	if err != nil {
		return err
	}
	k.tlsConf = tlsConf
	return nil
}

func (k *kafkaClient) configureWriter(tlsConf *tls.Config) (*kafka.Writer, error) {
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = defaultKafkaTopic
	}
	if tlsConf == nil || tlsConf.InsecureSkipVerify {
		return &kafka.Writer{
			Addr:     kafka.TCP(k.address),
			Topic:    kafkaTopic,
			Balancer: &kafka.LeastBytes{},
		}, nil
	}
	if err := k.watchWriter(); err != nil {
		return nil, err
	}
	return &kafka.Writer{
		Addr:     kafka.TCP(k.address),
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{},
		Transport: &kafka.Transport{
			TLS: tlsConf,
		},
	}, nil
}

func (k *kafkaClient) watchWriter() error {
	err := k.tm.WatchKafkaTLSConf(func(conf *tls.Config, err error) {
		if err != nil {
			log.Fatalf("an error occurred in the watch %s client: %s", tlsmanager.ClientKafka, err.Error())
		}
		k.tlsConf = conf
		if err != nil {
			log.Fatal("cannot apply new kafka tls configuration, exiting: ", err.Error())
		}
	})
	return err
}
