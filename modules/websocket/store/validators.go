package store

import (
	"context"
	"fmt"
	"sync"

	"gitlab.com/jideobs/nebularcore/modules/websocket/connections"
	"gitlab.com/jideobs/nebularcore/modules/websocket/protocol"
)

// TopicValidatorFunc is called before a subscribe or publish is processed.
// conn is the authenticated connection; topic is the full topic string from the client message.
type TopicValidatorFunc func(ctx context.Context, conn connections.Connection, topic string) error

// registeredValidator pairs a glob pattern with a validator function.
type registeredValidator struct {
	pattern string
	fn      TopicValidatorFunc
}

// ValidatorRegistry stores topic validators and executes matching ones during
// subscribe and publish operations. It is safe for concurrent reads; writes
// (registration) are expected only at module initialisation time.
type ValidatorRegistry struct {
	mu         sync.RWMutex
	validators []registeredValidator
}

// NewValidatorRegistry creates an empty ValidatorRegistry.
func NewValidatorRegistry() *ValidatorRegistry {
	return &ValidatorRegistry{}
}

// Register adds a validator for topics matching the given glob pattern.
// The pattern must be non-empty and uses the same syntax as MatchTopic
func (r *ValidatorRegistry) Register(pattern string, fn TopicValidatorFunc) error {
	if pattern == "" {
		return fmt.Errorf("topic validator pattern must not be empty")
	}
	if fn == nil {
		return fmt.Errorf("topic validator function must not be nil")
	}
	r.mu.Lock()
	r.validators = append(r.validators, registeredValidator{pattern: pattern, fn: fn})
	r.mu.Unlock()
	return nil
}

// Validate runs all registered validators whose pattern matches the given topic.
// If any validator returns a non-nil error the operation is rejected immediately.
// If no validators match the topic, nil is returned (allow by default).
func (r *ValidatorRegistry) Validate(ctx context.Context, conn connections.Connection, topic string) error {
	r.mu.RLock()
	validators := r.validators
	r.mu.RUnlock()

	for _, v := range validators {
		if protocol.MatchTopic(v.pattern, topic) {
			if err := v.fn(ctx, conn, topic); err != nil {
				return err
			}
		}
	}
	return nil
}
