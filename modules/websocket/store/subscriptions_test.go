package store

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubscribeAndGetTopics(t *testing.T) {
	s := NewSubscriptions()

	s.Subscribe("conn1", "chat.room1")
	s.Subscribe("conn1", "chat.room2")

	topics := s.GetTopics("conn1")
	sort.Strings(topics)
	assert.Equal(t, []string{"chat.room1", "chat.room2"}, topics)
}

func TestUnsubscribe(t *testing.T) {
	s := NewSubscriptions()

	s.Subscribe("conn1", "chat.room1")
	s.Subscribe("conn1", "chat.room2")
	s.Unsubscribe("conn1", "chat.room1")

	topics := s.GetTopics("conn1")
	assert.Equal(t, []string{"chat.room2"}, topics)

	assert.Equal(t, 0, s.TopicSubscriberCount("chat.room1"))
	assert.Equal(t, 1, s.TopicSubscriberCount("chat.room2"))
}

func TestUnsubscribeAll(t *testing.T) {
	s := NewSubscriptions()

	s.Subscribe("conn1", "topic1")
	s.Subscribe("conn1", "topic2")
	s.Subscribe("conn2", "topic1")

	s.UnsubscribeAll("conn1")

	assert.Nil(t, s.GetTopics("conn1"))
	assert.Equal(t, 0, s.TopicSubscriberCount("topic2"))
	assert.Equal(t, 1, s.TopicSubscriberCount("topic1"))
}

func TestGetSubscribedConns(t *testing.T) {
	s := NewSubscriptions()

	s.Subscribe("conn1", "chat.*")
	s.Subscribe("conn2", "chat.room1")
	s.Subscribe("conn3", "other.topic")

	conns := s.GetSubscribedConns("chat.room1")
	sort.Strings(conns)
	assert.Equal(t, []string{"conn1", "conn2"}, conns)

	conns = s.GetSubscribedConns("other.topic")
	assert.Equal(t, []string{"conn3"}, conns)
}

func TestGetTopicsNoSubscriptions(t *testing.T) {
	s := NewSubscriptions()
	assert.Nil(t, s.GetTopics("nonexistent"))
}

func TestTopicSubscriberCountNoSubscribers(t *testing.T) {
	s := NewSubscriptions()
	assert.Equal(t, 0, s.TopicSubscriberCount("nonexistent"))
}

func TestDuplicateSubscribe(t *testing.T) {
	s := NewSubscriptions()

	s.Subscribe("conn1", "topic1")
	s.Subscribe("conn1", "topic1") // duplicate

	topics := s.GetTopics("conn1")
	assert.Equal(t, 1, len(topics))
	assert.Equal(t, 1, s.TopicSubscriberCount("topic1"))
}
