package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithRetry(t *testing.T) {
	tests := []struct {
		name       string
		handler    Handler
		config     RetryConfig
		wantErr    bool
		wantCalls  int
		wantErrMsg string
	}{
		{
			name: "successful first try",
			handler: func(ctx context.Context, e Message) error {
				return nil
			},
			config: RetryConfig{
				MaxRetries:  3,
				InitialWait: time.Millisecond,
				MaxWait:     time.Millisecond * 10,
			},
			wantCalls: 1,
		},
		{
			name: "succeeds after retry",
			handler: func() Handler {
				calls := 0
				return func(ctx context.Context, e Message) error {
					calls++
					if calls < 2 {
						return errors.New("temporary error")
					}
					return nil
				}
			}(),
			config: RetryConfig{
				MaxRetries:  3,
				InitialWait: time.Millisecond,
				MaxWait:     time.Millisecond * 10,
			},
			wantCalls: 2,
		},
		{
			name: "fails after max retries",
			handler: func(ctx context.Context, e Message) error {
				return errors.New("persistent error")
			},
			config: RetryConfig{
				MaxRetries:  2,
				InitialWait: time.Millisecond,
				MaxWait:     time.Millisecond * 10,
			},
			wantErr:    true,
			wantCalls:  2,
			wantErrMsg: "max retries reached: persistent error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			wrapped := WithRetry(func(ctx context.Context, e Message) error {
				calls++
				return tt.handler(ctx, e)
			}, tt.config)

			// Create a properly initialized test event
			testEvent, err := NewMessage("test.event", "test-source", "test payload")
			assert.NoError(t, err)

			err = wrapped(context.Background(), testEvent)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Contains(t, err.Error(), tt.wantErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantCalls, calls)
		})
	}
}

func TestWithTimeout(t *testing.T) {
	tests := []struct {
		name    string
		handler Handler
		timeout time.Duration
		wantErr bool
	}{
		{
			name: "completes within timeout",
			handler: func(ctx context.Context, e Message) error {
				time.Sleep(time.Millisecond)
				return nil
			},
			timeout: time.Second,
		},
		{
			name: "exceeds timeout",
			handler: func(ctx context.Context, e Message) error {
				time.Sleep(time.Millisecond * 100)
				return nil
			},
			timeout: time.Millisecond * 10,
			wantErr: true,
		},
		{
			name: "respects context cancellation",
			handler: func(ctx context.Context, e Message) error {
				<-ctx.Done()
				return ctx.Err()
			},
			timeout: time.Millisecond * 10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WithTimeout(tt.handler, tt.timeout)
			// Create a properly initialized test event
			testEvent, err := NewMessage("test.event", "test-source", "test payload")
			assert.NoError(t, err)

			err = wrapped(context.Background(), testEvent)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWithRecovery(t *testing.T) {
	tests := []struct {
		name    string
		handler Handler
		wantErr bool
	}{
		{
			name: "no panic",
			handler: func(ctx context.Context, e Message) error {
				return nil
			},
		},
		{
			name: "recovers from panic",
			handler: func(ctx context.Context, e Message) error {
				panic("something went wrong")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := WithRecovery(tt.handler)
			// Create a properly initialized test event
			testEvent, err := NewMessage("test.event", "test-source", "test payload")
			assert.NoError(t, err)

			err = wrapped(context.Background(), testEvent)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "handler panic")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestChain(t *testing.T) {
	tests := []struct {
		name     string
		handlers []Handler
		wantErr  bool
	}{
		{
			name:     "empty chain",
			handlers: []Handler{},
		},
		{
			name: "single handler",
			handlers: []Handler{
				func(ctx context.Context, e Message) error {
					return nil
				},
			},
		},
		{
			name: "multiple successful handlers",
			handlers: []Handler{
				func(ctx context.Context, e Message) error {
					return nil
				},
				func(ctx context.Context, e Message) error {
					return nil
				},
			},
		},
		{
			name: "stops on first error",
			handlers: []Handler{
				func(ctx context.Context, e Message) error {
					return nil
				},
				func(ctx context.Context, e Message) error {
					return errors.New("handler error")
				},
				func(ctx context.Context, e Message) error {
					return nil
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chained := Chain(tt.handlers...)
			err := chained(context.Background(), Message{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
