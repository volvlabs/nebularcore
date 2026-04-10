package emitter

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/volvlabs/nebularcore/modules/auth/models"
	"github.com/volvlabs/nebularcore/modules/event"
	"github.com/volvlabs/nebularcore/modules/event/mocks"
)

func TestEventEmitter_EmitAuthEvent(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() (EventEmitter, *mocks.EventBus, *models.User)
		data    AuthEventData
		wantErr bool
	}{
		{
			name: "successful auth event emission",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(nil)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			data: AuthEventData{
				UserID:    "test-user-id",
				Email:     "test@example.com",
				IP:        "127.0.0.1",
				UserAgent: "test-agent",
				Timestamp: time.Now(),
				EventType: EventLoginSuccess,
				Success:   true,
			},
			wantErr: false,
		},
		{
			name: "event bus publish error",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(assert.AnError)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			data: AuthEventData{
				UserID:    "test-user-id",
				Email:     "test@example.com",
				EventType: EventLoginSuccess,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter, eventBus, user := tt.setup()
			err := emitter.EmitAuthEvent(context.Background(), user, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			eventBus.AssertExpectations(t)
		})
	}
}

func TestEventEmitter_EmitLoginEvent(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (EventEmitter, *mocks.EventBus, *models.User)
		ip        string
		userAgent string
		success   bool
		wantErr   bool
	}{
		{
			name: "successful login event",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(nil)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			ip:        "127.0.0.1",
			userAgent: "test-agent",
			success:   true,
			wantErr:   false,
		},
		{
			name: "failed login event",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(nil)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			ip:        "127.0.0.1",
			userAgent: "test-agent",
			success:   false,
			wantErr:   false,
		},
		{
			name: "event bus error",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(assert.AnError)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			ip:        "127.0.0.1",
			userAgent: "test-agent",
			success:   true,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter, eventBus, user := tt.setup()
			err := emitter.EmitLoginEvent(context.Background(), user, tt.ip, tt.userAgent, tt.success)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			eventBus.AssertExpectations(t)
		})
	}
}

func TestEventEmitter_EmitPasswordEvent(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (EventEmitter, *mocks.EventBus, *models.User)
		ip        string
		userAgent string
		eventType string
		wantErr   bool
	}{
		{
			name: "successful password reset event",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(nil)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			ip:        "127.0.0.1",
			userAgent: "test-agent",
			eventType: EventPasswordReset,
			wantErr:   false,
		},
		{
			name: "successful password change event",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(nil)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			ip:        "127.0.0.1",
			userAgent: "test-agent",
			eventType: EventPasswordChanged,
			wantErr:   false,
		},
		{
			name: "event bus error",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(assert.AnError)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			ip:        "127.0.0.1",
			userAgent: "test-agent",
			eventType: EventPasswordReset,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter, eventBus, user := tt.setup()
			err := emitter.EmitPasswordEvent(context.Background(), user, tt.ip, tt.userAgent, tt.eventType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			eventBus.AssertExpectations(t)
		})
	}
}

func TestEventEmitter_EmitUserEvent(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() (EventEmitter, *mocks.EventBus, *models.User)
		eventType string
		wantErr   bool
	}{
		{
			name: "successful user created event",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(nil)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			eventType: EventUserCreated,
			wantErr:   false,
		},
		{
			name: "successful user updated event",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(nil)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			eventType: EventUserUpdated,
			wantErr:   false,
		},
		{
			name: "event bus error",
			setup: func() (EventEmitter, *mocks.EventBus, *models.User) {
				eventBus := mocks.NewEventBus(t)
				eventBus.On("Publish", mock.MatchedBy(func(ctx context.Context) bool { return true }), mock.MatchedBy(func(e event.Message) bool { return true })).Return(assert.AnError)
				return NewEventEmitter(eventBus), eventBus, &models.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}
			},
			eventType: EventUserCreated,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emitter, eventBus, user := tt.setup()
			err := emitter.EmitUserEvent(context.Background(), user, tt.eventType)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			eventBus.AssertExpectations(t)
		})
	}
}
