package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/config"
	"github.com/sw5005-sus/ceramicraft-commodity-mservice/server/log"
)

const TopicProductChanged = "product_changed"

type ProductChangedEvent struct {
	ProductName string `json:"productName"`
	Operation   string `json:"operation"`
}

var (
	producer     *KafkaProducer
	producerOnce sync.Once
)

type KafkaProducer struct {
	writer *kafka.Writer
}

func InitProducer() {
	producerOnce.Do(func() {
		producer = &KafkaProducer{
			writer: &kafka.Writer{
				Addr:                   kafka.TCP(config.Config.KafkaConfig.Brokers...),
				Balancer:               &kafka.LeastBytes{},
				AllowAutoTopicCreation: true,
			},
		}
		log.Logger.Info("Kafka producer initialized")
	})
}

func CloseProducer() {
	if producer != nil && producer.writer != nil {
		if err := producer.writer.Close(); err != nil {
			log.Logger.Errorf("Failed to close Kafka producer: %v", err)
		}
	}
}

func GetProducer() *KafkaProducer {
	return producer
}

func (p *KafkaProducer) PublishProductChanged(ctx context.Context, productName, operation string) error {
	if p == nil || p.writer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}
	event := ProductChangedEvent{
		ProductName: productName,
		Operation:   operation,
	}
	value, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal product changed event: %w", err)
	}

	err = p.writer.WriteMessages(ctx, kafka.Message{
		Topic: TopicProductChanged,
		Key:   []byte(productName),
		Value: value,
	})
	if err != nil {
		log.Logger.Errorf("Failed to publish product_changed event: product=%s, op=%s, err=%v", productName, operation, err)
		return err
	}
	log.Logger.Infof("Published product_changed event: product=%s, op=%s", productName, operation)
	return nil
}
