package protocol

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Encode serializes a ServerMessage to JSON bytes.
func Encode(msg *ServerMessage) ([]byte, error) {
	return json.Marshal(msg)
}

// NewSubscribedMsg creates a subscription confirmation message.
func NewSubscribedMsg(requestID, topic string) *ServerMessage {
	return &ServerMessage{
		ID:        requestID,
		Type:      TypeSubscribed,
		Topic:     topic,
		Timestamp: time.Now().UTC(),
	}
}

// NewUnsubscribedMsg creates an unsubscription confirmation message.
func NewUnsubscribedMsg(requestID, topic string) *ServerMessage {
	return &ServerMessage{
		ID:        requestID,
		Type:      TypeUnsubscribed,
		Topic:     topic,
		Timestamp: time.Now().UTC(),
	}
}

// NewEventMsg creates a server-to-client event message.
func NewEventMsg(topic string, payload json.RawMessage) *ServerMessage {
	return &ServerMessage{
		ID:        uuid.NewString(),
		Type:      TypeMessage,
		Topic:     topic,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}

// NewErrorMsg creates an error response message.
func NewErrorMsg(requestID, errText string) *ServerMessage {
	payload, _ := json.Marshal(map[string]string{"error": errText})
	return &ServerMessage{
		ID:        requestID,
		Type:      TypeError,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
}

// NewPongMsg creates a pong response message.
func NewPongMsg(requestID string) *ServerMessage {
	return &ServerMessage{
		ID:        requestID,
		Type:      TypePong,
		Timestamp: time.Now().UTC(),
	}
}

// NewAuthSuccessMsg creates an auth success response message.
func NewAuthSuccessMsg(requestID string) *ServerMessage {
	return &ServerMessage{
		ID:        requestID,
		Type:      TypeAuthSuccess,
		Timestamp: time.Now().UTC(),
	}
}
