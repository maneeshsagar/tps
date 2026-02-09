package ports

import "context"

type Message struct {
	Key   string
	Value []byte
}

// MessageProducer defines the interface for publishing messages to a kafka.
type MessageProducer interface {
	Publish(ctx context.Context, topic string, msg Message) error
	Close() error
}

// MessageConsumer defines the interface for subscribing to messages from a kafka.
type MessageConsumer interface {
	Subscribe(ctx context.Context, topic string, handler func(msg Message) error) error
	Close() error
}
