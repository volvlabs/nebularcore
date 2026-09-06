package protocol

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_ValidSubscribe(t *testing.T) {
	raw := `{"id":"1","type":"subscribe","topic":"notifications.user:123","timestamp":"2026-01-01T00:00:00Z"}`
	msg, err := Parse([]byte(raw), 256)
	require.NoError(t, err)
	assert.Equal(t, TypeSubscribe, msg.Type)
	assert.Equal(t, "notifications.user:123", msg.Topic)
	assert.Equal(t, "1", msg.ID)
}

func TestParse_ValidPublish(t *testing.T) {
	raw := `{"id":"2","type":"publish","topic":"chat.room:1","payload":{"text":"hello"},"timestamp":"2026-01-01T00:00:00Z"}`
	msg, err := Parse([]byte(raw), 256)
	require.NoError(t, err)
	assert.Equal(t, TypePublish, msg.Type)
	assert.Equal(t, "chat.room:1", msg.Topic)
	assert.NotNil(t, msg.Payload)
}

func TestParse_ValidPing(t *testing.T) {
	raw := `{"id":"3","type":"ping","timestamp":"2026-01-01T00:00:00Z"}`
	msg, err := Parse([]byte(raw), 256)
	require.NoError(t, err)
	assert.Equal(t, TypePing, msg.Type)
}

func TestParse_ValidUnsubscribe(t *testing.T) {
	raw := `{"id":"4","type":"unsubscribe","topic":"notifications.*","timestamp":"2026-01-01T00:00:00Z"}`
	msg, err := Parse([]byte(raw), 256)
	require.NoError(t, err)
	assert.Equal(t, TypeUnsubscribe, msg.Type)
}

func TestParse_ValidAuth(t *testing.T) {
	raw := `{"id":"5","type":"auth","payload":{"token":"abc"},"timestamp":"2026-01-01T00:00:00Z"}`
	msg, err := Parse([]byte(raw), 256)
	require.NoError(t, err)
	assert.Equal(t, TypeAuth, msg.Type)
}

func TestParse_MalformedJSON(t *testing.T) {
	_, err := Parse([]byte(`{not json`), 256)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON")
}

func TestParse_MissingType(t *testing.T) {
	raw := `{"id":"1","topic":"test"}`
	_, err := Parse([]byte(raw), 256)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing message type")
}

func TestParse_UnknownType(t *testing.T) {
	raw := `{"id":"1","type":"unknown","topic":"test"}`
	_, err := Parse([]byte(raw), 256)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown message type")
}

func TestParse_MissingTopicForSubscribe(t *testing.T) {
	raw := `{"id":"1","type":"subscribe"}`
	_, err := Parse([]byte(raw), 256)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "topic is required")
}

func TestParse_TopicExceedsMaxLength(t *testing.T) {
	longTopic := make([]byte, 300)
	for i := range longTopic {
		longTopic[i] = 'a'
	}
	raw, _ := json.Marshal(Message{ID: "1", Type: TypeSubscribe, Topic: string(longTopic)})
	_, err := Parse(raw, 256)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "topic exceeds max length")
}

func TestParse_ZeroMaxTopicLengthSkipsCheck(t *testing.T) {
	longTopic := make([]byte, 1000)
	for i := range longTopic {
		longTopic[i] = 'a'
	}
	raw, _ := json.Marshal(Message{ID: "1", Type: TypeSubscribe, Topic: string(longTopic)})
	msg, err := Parse(raw, 0)
	require.NoError(t, err)
	assert.Equal(t, TypeSubscribe, msg.Type)
}

func TestEncode_RoundTrip(t *testing.T) {
	original := NewEventMsg("test.topic", json.RawMessage(`{"key":"value"}`))
	data, err := Encode(original)
	require.NoError(t, err)

	var decoded ServerMessage
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.Topic, decoded.Topic)
	assert.Equal(t, original.ID, decoded.ID)
}

func TestNewErrorMsg(t *testing.T) {
	msg := NewErrorMsg("req-1", "something went wrong")
	assert.Equal(t, TypeError, msg.Type)
	assert.Equal(t, "req-1", msg.ID)
	assert.Contains(t, string(msg.Payload), "something went wrong")
}

func TestNewPongMsg(t *testing.T) {
	msg := NewPongMsg("req-2")
	assert.Equal(t, TypePong, msg.Type)
	assert.Equal(t, "req-2", msg.ID)
}

func TestNewSubscribedMsg(t *testing.T) {
	msg := NewSubscribedMsg("req-3", "topic.a")
	assert.Equal(t, TypeSubscribed, msg.Type)
	assert.Equal(t, "topic.a", msg.Topic)
}

func TestNewUnsubscribedMsg(t *testing.T) {
	msg := NewUnsubscribedMsg("req-4", "topic.b")
	assert.Equal(t, TypeUnsubscribed, msg.Type)
	assert.Equal(t, "topic.b", msg.Topic)
}

func TestNewAuthSuccessMsg(t *testing.T) {
	msg := NewAuthSuccessMsg("req-5")
	assert.Equal(t, TypeAuthSuccess, msg.Type)
	assert.Equal(t, "req-5", msg.ID)
}
