package mq

import (
	"context"
	"time"

	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
	"github.com/segmentio/kafka-go"
)

type KafkaMsgProcessor func(msg []byte) error

const (
	topicOrderCreated = "order_created"
)

func Init() {
	InitProducer()
	startKafkaConsumer(topicOrderCreated, clearCartProcess)
	log.Logger.Infof("Kafka consumer for topic %s started", topicOrderCreated)
}

func Close() {
	CloseProducer()
}

func startKafkaConsumer(topic string, processor KafkaMsgProcessor) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        config.Config.KafkaConfig.Brokers,
		Topic:          topic,
		GroupID:        config.Config.KafkaConfig.GroupID,
		MaxBytes:       config.Config.KafkaConfig.MaxBytes,
		CommitInterval: time.Duration(config.Config.KafkaConfig.CommitInterval),
	})
	go func() {
		for {
			ctx := context.Background()
			m, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Logger.Errorf("Error reading message: %v", err)
				continue
			}
			log.Logger.Infof("Message received: Topic=%s, Key=%s, Value=%s", m.Topic, m.Key, string(m.Value))
			err = processor(m.Value)
			if err == nil {
				cmitErr := reader.CommitMessages(ctx, m)
				if cmitErr != nil {
					log.Logger.Errorf("Failed to commit message at offset %d: %v", m.Offset, cmitErr)
				}
				log.Logger.Infof("Topic: %s, Key: %s, Message at offset %d processed and committed", m.Topic, m.Key, m.Offset)
			}
		}
	}()
}
