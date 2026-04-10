package protocol

import "strings"

// MatchTopic checks if a topic matches a pattern.
// Patterns support:
//   - exact match: "user.created" matches "user.created"
//   - wildcard segment: "user.*" matches "user.created", "user.deleted"
//   - multi-segment wildcard: "**" matches everything
//
// Segments are separated by ".".
func MatchTopic(pattern, topic string) bool {
	if pattern == "**" || pattern == topic {
		return true
	}

	patternParts := strings.Split(pattern, ".")
	topicParts := strings.Split(topic, ".")

	return matchParts(patternParts, topicParts)
}

func matchParts(pattern, topic []string) bool {
	pi, ti := 0, 0
	for pi < len(pattern) && ti < len(topic) {
		if pattern[pi] == "**" {
			// ** matches zero or more segments.
			if pi == len(pattern)-1 {
				return true
			}
			// Try matching the rest of the pattern at each position.
			for k := ti; k <= len(topic); k++ {
				if matchParts(pattern[pi+1:], topic[k:]) {
					return true
				}
			}
			return false
		}
		if pattern[pi] == "*" {
			// * matches exactly one segment.
			pi++
			ti++
			continue
		}
		if pattern[pi] != topic[ti] {
			return false
		}
		pi++
		ti++
	}
	return pi == len(pattern) && ti == len(topic)
}

// MatchAny returns true if the topic matches any of the patterns.
func MatchAny(patterns []string, topic string) bool {
	for _, p := range patterns {
		if MatchTopic(p, topic) {
			return true
		}
	}
	return false
}
