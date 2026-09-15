package bridge

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchTopic(t *testing.T) {
	tests := []struct {
		pattern string
		topic   string
		want    bool
	}{
		// Exact match.
		{"user.created", "user.created", true},
		{"user.created", "user.deleted", false},

		// Single wildcard.
		{"user.*", "user.created", true},
		{"user.*", "user.deleted", true},
		{"user.*", "user.profile.updated", false},
		{"*.created", "user.created", true},
		{"*.created", "order.created", true},

		// Double wildcard.
		{"**", "user.created", true},
		{"**", "a.b.c.d", true},
		{"user.**", "user.created", true},
		{"user.**", "user.profile.updated", true},
		{"user.**", "order.created", false},

		// Mixed.
		{"user.*.updated", "user.profile.updated", true},
		{"user.*.updated", "user.profile.deleted", false},

		// Edge cases.
		{"", "", true},
		{"user", "user", true},
		{"user", "order", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.topic, func(t *testing.T) {
			assert.Equal(t, tt.want, MatchTopic(tt.pattern, tt.topic))
		})
	}
}

func TestMatchAny(t *testing.T) {
	patterns := []string{"user.*", "notification.**"}

	assert.True(t, MatchAny(patterns, "user.created"))
	assert.True(t, MatchAny(patterns, "notification.email.sent"))
	assert.False(t, MatchAny(patterns, "order.created"))
}

func TestMatchAny_Empty(t *testing.T) {
	assert.False(t, MatchAny(nil, "anything"))
	assert.False(t, MatchAny([]string{}, "anything"))
}
