package bridge

import "gitlab.com/jideobs/nebularcore/modules/websocket/protocol"

// MatchTopic delegates to protocol.MatchTopic.
func MatchTopic(pattern, topic string) bool {
	return protocol.MatchTopic(pattern, topic)
}

// MatchAny delegates to protocol.MatchAny.
func MatchAny(patterns []string, topic string) bool {
	return protocol.MatchAny(patterns, topic)
}
