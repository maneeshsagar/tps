package infrastructure

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"github.com/maneeshsagar/tps/internal/core/ports"
	"github.com/maneeshsagar/tps/logger"
)

// Producer

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		panic("kafka producer: " + err.Error())
	}
	return &KafkaProducer{producer}
}

func (p *KafkaProducer) Publish(ctx context.Context, topic string, msg ports.Message) error {
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(msg.Key),
		Value: sarama.ByteEncoder(msg.Value),
	})
	return err
}

func (p *KafkaProducer) Close() error {
	return p.producer.Close()
}

// Consumer

type KafkaConsumer struct {
	brokers []string
	log     logger.Logger
}

func NewKafkaConsumer(brokers []string, log logger.Logger) *KafkaConsumer {
	return &KafkaConsumer{brokers, log}
}

func (c *KafkaConsumer) Subscribe(ctx context.Context, topic string, handler func(msg ports.Message) error) error {
	cfg := sarama.NewConfig()
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Offsets.AutoCommit.Enable = false

	group, err := sarama.NewConsumerGroup(c.brokers, "tps-consumer", cfg)
	if err != nil {
		return err
	}
	defer group.Close()

	h := &consumerHandler{handler: handler, log: c.log}

	c.log.Info("consumer joined group", "topic", topic)

	for {
		if err := group.Consume(ctx, []string{topic}, h); err != nil {
			c.log.Error("consume error", "err", err)
			time.Sleep(time.Second)
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

func (c *KafkaConsumer) Close() error {
	return nil
}

// consumerHandler implements sarama.ConsumerGroupHandler
type consumerHandler struct {
	handler func(msg ports.Message) error
	log     logger.Logger
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.log.Info("received message", "topic", msg.Topic, "partition", msg.Partition, "offset", msg.Offset)

		err := h.handler(ports.Message{
			Key:   string(msg.Key),
			Value: msg.Value,
		})
		if err != nil {
			h.log.Error("handler failed, skipping commit", "err", err)
			continue
		}

		// commit only after successful processing
		session.MarkMessage(msg, "")
		session.Commit()
	}
	return nil
}
