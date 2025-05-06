package protocol

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of a WebSocket message.
type MessageType string

// Client-to-server message types.
const (
	TypeSubscribe   MessageType = "subscribe"
	TypePublish     MessageType = "publish"
	TypeUnsubscribe MessageType = "unsubscribe"
	TypePing        MessageType = "ping"
	TypeAuth        MessageType = "auth"
)

// Server-to-client message types.
const (
	TypeSubscribed   MessageType = "subscribed"
	TypeUnsubscribed MessageType = "unsubscribed"
	TypeMessage      MessageType = "message"
	TypeError        MessageType = "error"
	TypePong         MessageType = "pong"
	TypeAuthSuccess  MessageType = "auth_success"
)

// Message represents a client-to-server WebSocket message.
type Message struct {
	ID        string          `json:"id"`
	Type      MessageType     `json:"type"`
	Topic     string          `json:"topic,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// ServerMessage represents a server-to-client WebSocket message.
type ServerMessage struct {
	ID        string          `json:"id"`
	Type      MessageType     `json:"type"`
	Topic     string          `json:"topic,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}
