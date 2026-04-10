package store

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/volvlabs/nebularcore/modules/websocket/connections"
	"github.com/volvlabs/nebularcore/modules/websocket/protocol"
)

// mockConn is a minimal Connection implementation for testing.
type mockConn struct {
	id       string
	userID   string
	tenantID string
	sent     []*protocol.ServerMessage
}

func (m *mockConn) ID() string                            { return m.id }
func (m *mockConn) UserID() string                        { return m.userID }
func (m *mockConn) TenantID() string                      { return m.tenantID }
func (m *mockConn) Send(msg *protocol.ServerMessage) bool { m.sent = append(m.sent, msg); return true }
func (m *mockConn) Close()                                {}
func (m *mockConn) Context() context.Context              { return context.Background() }

func newMockConn(id, userID string) connections.Connection {
	return &mockConn{id: id, userID: userID}
}

func TestValidatorRegistry_NoValidators(t *testing.T) {
	r := NewValidatorRegistry()
	conn := newMockConn("conn1", "user1")

	err := r.Validate(context.Background(), conn, "any.topic.here")
	assert.NoError(t, err, "should allow when no validators are registered")
}

func TestValidatorRegistry_MatchingValidatorAllows(t *testing.T) {
	r := NewValidatorRegistry()
	conn := newMockConn("conn1", "user1")

	err := r.Register("qa.user.*", func(ctx context.Context, c connections.Connection, topic string) error {
		return nil
	})
	require.NoError(t, err)

	err = r.Validate(context.Background(), conn, "qa.user.123")
	assert.NoError(t, err, "should allow when matching validator returns nil")
}

func TestValidatorRegistry_MatchingValidatorRejects(t *testing.T) {
	r := NewValidatorRegistry()
	conn := newMockConn("conn1", "user1")

	err := r.Register("qa.user.*", func(ctx context.Context, c connections.Connection, topic string) error {
		return errors.New("not your topic")
	})
	require.NoError(t, err)

	err = r.Validate(context.Background(), conn, "qa.user.123")
	assert.EqualError(t, err, "not your topic")
}

func TestValidatorRegistry_NonMatchingValidatorIgnored(t *testing.T) {
	r := NewValidatorRegistry()
	conn := newMockConn("conn1", "user1")

	called := false
	err := r.Register("qa.conversation.*", func(ctx context.Context, c connections.Connection, topic string) error {
		called = true
		return errors.New("should not be called")
	})
	require.NoError(t, err)

	err = r.Validate(context.Background(), conn, "chat.room.456")
	assert.NoError(t, err, "should allow when no validators match the topic")
	assert.False(t, called, "non-matching validator should not be called")
}

func TestValidatorRegistry_MultipleValidatorsOneRejects(t *testing.T) {
	r := NewValidatorRegistry()
	conn := newMockConn("conn1", "user1")

	err := r.Register("qa.**", func(ctx context.Context, c connections.Connection, topic string) error {
		return nil // allow
	})
	require.NoError(t, err)

	err = r.Register("qa.user.*", func(ctx context.Context, c connections.Connection, topic string) error {
		return errors.New("access denied")
	})
	require.NoError(t, err)

	err = r.Validate(context.Background(), conn, "qa.user.123")
	assert.EqualError(t, err, "access denied", "should reject if any validator rejects")
}

func TestValidatorRegistry_MultipleValidatorsAllAllow(t *testing.T) {
	r := NewValidatorRegistry()
	conn := newMockConn("conn1", "user1")

	callOrder := []string{}

	err := r.Register("qa.**", func(ctx context.Context, c connections.Connection, topic string) error {
		callOrder = append(callOrder, "first")
		return nil
	})
	require.NoError(t, err)

	err = r.Register("qa.user.*", func(ctx context.Context, c connections.Connection, topic string) error {
		callOrder = append(callOrder, "second")
		return nil
	})
	require.NoError(t, err)

	err = r.Validate(context.Background(), conn, "qa.user.123")
	assert.NoError(t, err)
	assert.Equal(t, []string{"first", "second"}, callOrder, "all matching validators should be called in order")
}

func TestValidatorRegistry_RegisterEmptyPattern(t *testing.T) {
	r := NewValidatorRegistry()
	err := r.Register("", func(ctx context.Context, c connections.Connection, topic string) error {
		return nil
	})
	assert.Error(t, err, "should reject empty pattern")
}

func TestValidatorRegistry_RegisterNilFunc(t *testing.T) {
	r := NewValidatorRegistry()
	err := r.Register("qa.*", nil)
	assert.Error(t, err, "should reject nil validator function")
}

func TestValidatorRegistry_ValidatorReceivesCorrectArgs(t *testing.T) {
	r := NewValidatorRegistry()
	conn := newMockConn("conn1", "user42")

	var receivedTopic string
	var receivedUserID string

	err := r.Register("chat.*", func(ctx context.Context, c connections.Connection, topic string) error {
		receivedTopic = topic
		receivedUserID = c.UserID()
		return nil
	})
	require.NoError(t, err)

	err = r.Validate(context.Background(), conn, "chat.room1")
	assert.NoError(t, err)
	assert.Equal(t, "chat.room1", receivedTopic)
	assert.Equal(t, "user42", receivedUserID)
}
