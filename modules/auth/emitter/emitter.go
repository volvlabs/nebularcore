package emitter

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gitlab.com/jideobs/nebularcore/modules/auth/interfaces"
	"gitlab.com/jideobs/nebularcore/modules/event"
)

const (
	EventUserCreated            = "user.created"
	EventUserUpdated            = "user.updated"
	EventUserDeleted            = "user.deleted"
	EventLoginSuccess           = "auth.login.success"
	EventLoginFailed            = "auth.login.failed"
	EventPasswordReset          = "auth.password.reset"
	EventPasswordResetInitiated = "auth.password.reset.initiated"
	EventPasswordChanged        = "auth.password.changed"
)

type UserEventData struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

type LoginEventData struct {
	UserID    string    `json:"userId"`
	Email     string    `json:"email"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"userAgent"`
	Timestamp time.Time `json:"timestamp"`
}

type LoginFailedEventData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Timestamp time.Time `json:"timestamp"`
}

type AuthEventData struct {
	UserID      string    `json:"user_id"`
	Email       string    `json:"email"`
	PhoneNumber string    `json:"phoneNumber"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent"`
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"` // Type of auth event (login, password reset, etc.)
	Success     bool      `json:"success"`    // For events that can succeed/fail
}

type EventEmitter interface {
	EmitAuthEvent(ctx context.Context, user interfaces.User, data AuthEventData) error
	EmitLoginEvent(ctx context.Context, user interfaces.User, ip, userAgent string, success bool) error
	EmitPasswordEvent(ctx context.Context, user interfaces.User, ip, userAgent string, eventType string) error
	EmitUserEvent(ctx context.Context, user interfaces.User, eventType string) error
	EmitPasswordResetEvent(ctx context.Context, user interfaces.User) error
	EmitPasswordResetInitiatedEvent(ctx context.Context, user interfaces.User) error
	EmitPasswordChangedEvent(ctx context.Context, user interfaces.User) error
}

type eventEmitter struct {
	eventBus event.Bus
}

func NewEventEmitter(eventBus event.Bus) EventEmitter {
	return &eventEmitter{
		eventBus: eventBus,
	}
}

func (m *eventEmitter) EmitAuthEvent(ctx context.Context, user interfaces.User, data AuthEventData) error {
	if user != nil {
		data.UserID = user.GetID().String()
		data.Email = user.GetEmail()
		data.PhoneNumber = user.GetPhoneNumber()
	}
	evt, err := event.NewMessage(data.EventType, "auth.module", data)
	if err != nil {
		return fmt.Errorf("failed to create auth event: %w", err)
	}
	err = m.eventBus.Publish(ctx, evt)
	if err != nil {
		log.Err(err).Msgf("PasswordManager: error occurred emitting event %s", data.EventType)
	}
	return err
}

func (m *eventEmitter) EmitLoginEvent(ctx context.Context, user interfaces.User, ip, userAgent string, success bool) error {
	data := AuthEventData{
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		EventType: EventLoginSuccess,
		Success:   success,
	}
	if !success {
		data.EventType = EventLoginFailed
	}
	return m.EmitAuthEvent(ctx, user, data)
}

func (m *eventEmitter) EmitPasswordEvent(ctx context.Context, user interfaces.User, ip, userAgent string, eventType string) error {
	data := AuthEventData{
		IP:        ip,
		UserAgent: userAgent,
		Timestamp: time.Now(),
		EventType: eventType,
	}
	return m.EmitAuthEvent(ctx, user, data)
}

func (m *eventEmitter) EmitUserEvent(ctx context.Context, user interfaces.User, eventType string) error {
	data := AuthEventData{
		Timestamp: time.Now(),
		EventType: eventType,
	}
	return m.EmitAuthEvent(ctx, user, data)
}

func (m *eventEmitter) EmitPasswordResetEvent(ctx context.Context, user interfaces.User) error {
	data := AuthEventData{
		Timestamp: time.Now(),
		EventType: EventPasswordReset,
	}
	return m.EmitAuthEvent(ctx, user, data)
}

func (m *eventEmitter) EmitPasswordResetInitiatedEvent(ctx context.Context, user interfaces.User) error {
	data := AuthEventData{
		Timestamp: time.Now(),
		EventType: EventPasswordResetInitiated,
	}
	return m.EmitAuthEvent(ctx, user, data)
}

func (m *eventEmitter) EmitPasswordChangedEvent(ctx context.Context, user interfaces.User) error {
	data := AuthEventData{
		Timestamp: time.Now(),
		EventType: EventPasswordChanged,
	}
	return m.EmitAuthEvent(ctx, user, data)
}
