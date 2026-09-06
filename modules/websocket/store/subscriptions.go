package store

import (
	"sync"

	"github.com/volvlabs/nebularcore/modules/websocket/protocol"
)

// Subscriptions tracks which connections are subscribed to which topics.
// Thread-safe via sync.Map.
type Subscriptions struct {
	// connTopics: connID -> *sync.Map (topic -> struct{})
	connTopics sync.Map
	// topicConns: topic -> *sync.Map (connID -> struct{})
	topicConns sync.Map
}

// NewSubscriptions creates a new Subscriptions store.
func NewSubscriptions() *Subscriptions {
	return &Subscriptions{}
}

// Subscribe adds a topic subscription for a connection.
func (s *Subscriptions) Subscribe(connID, topic string) {
	// connTopics
	actual, _ := s.connTopics.LoadOrStore(connID, &sync.Map{})
	actual.(*sync.Map).Store(topic, struct{}{})

	// topicConns
	actual, _ = s.topicConns.LoadOrStore(topic, &sync.Map{})
	actual.(*sync.Map).Store(connID, struct{}{})
}

// Unsubscribe removes a topic subscription for a connection.
func (s *Subscriptions) Unsubscribe(connID, topic string) {
	if v, ok := s.connTopics.Load(connID); ok {
		v.(*sync.Map).Delete(topic)
	}
	if v, ok := s.topicConns.Load(topic); ok {
		v.(*sync.Map).Delete(connID)
	}
}

// UnsubscribeAll removes all subscriptions for a connection.
func (s *Subscriptions) UnsubscribeAll(connID string) {
	v, ok := s.connTopics.Load(connID)
	if !ok {
		return
	}
	v.(*sync.Map).Range(func(key, _ any) bool {
		topic := key.(string)
		if tv, ok := s.topicConns.Load(topic); ok {
			tv.(*sync.Map).Delete(connID)
		}
		return true
	})
	s.connTopics.Delete(connID)
}

// GetTopics returns all topics a connection is subscribed to.
func (s *Subscriptions) GetTopics(connID string) []string {
	v, ok := s.connTopics.Load(connID)
	if !ok {
		return nil
	}
	var topics []string
	v.(*sync.Map).Range(func(key, _ any) bool {
		topics = append(topics, key.(string))
		return true
	})
	return topics
}

// GetSubscribedConns returns all connection IDs that should receive a message
// for the given event type. It checks each connection's subscribed topics
// using glob matching.
func (s *Subscriptions) GetSubscribedConns(eventType string) []string {
	var result []string
	s.connTopics.Range(func(connIDKey, topicsVal any) bool {
		connID := connIDKey.(string)
		topicsVal.(*sync.Map).Range(func(topicKey, _ any) bool {
			pattern := topicKey.(string)
			if protocol.MatchTopic(pattern, eventType) {
				result = append(result, connID)
				return false // found a match, no need to check more topics for this conn
			}
			return true
		})
		return true
	})
	return result
}

// TopicSubscriberCount returns the number of connections subscribed to a topic
// (exact match, not glob).
func (s *Subscriptions) TopicSubscriberCount(topic string) int {
	v, ok := s.topicConns.Load(topic)
	if !ok {
		return 0
	}
	count := 0
	v.(*sync.Map).Range(func(_, _ any) bool {
		count++
		return true
	})
	return count
}
