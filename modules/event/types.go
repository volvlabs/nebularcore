package event

import (
	"context"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
)

// Message represents a single event in the system
type Message struct {
	*message.Message

	Source    string
	EventType string
	Timestamp time.Time
}

type Handler func(context.Context, Message) error

type Publisher interface {
	Publish(ctx context.Context, event Message) error
	PublishAsync(ctx context.Context, event Message) (<-chan error, error)
}

type Subscriber interface {
	Subscribe(eventType string, handler Handler) error
	Unsubscribe(eventType string) error
}

type Bus interface {
	Publisher
	Subscriber
}
