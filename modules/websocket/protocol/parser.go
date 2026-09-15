package protocol

import (
	"encoding/json"
	"fmt"
)

// validClientTypes is the set of message types a client may send.
var validClientTypes = map[MessageType]bool{
	TypeSubscribe:   true,
	TypePublish:     true,
	TypeUnsubscribe: true,
	TypePing:        true,
	TypeAuth:        true,
}

// Parse decodes raw JSON bytes into a client Message and validates it.
func Parse(data []byte, maxTopicLength int) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	if msg.Type == "" {
		return nil, fmt.Errorf("missing message type")
	}
	if !validClientTypes[msg.Type] {
		return nil, fmt.Errorf("unknown message type: %s", msg.Type)
	}

	// Topic is required for subscribe, publish, unsubscribe.
	switch msg.Type {
	case TypeSubscribe, TypePublish, TypeUnsubscribe:
		if msg.Topic == "" {
			return nil, fmt.Errorf("topic is required for type %s", msg.Type)
		}
		if maxTopicLength > 0 && len(msg.Topic) > maxTopicLength {
			return nil, fmt.Errorf("topic exceeds max length of %d", maxTopicLength)
		}
	}

	return &msg, nil
}
