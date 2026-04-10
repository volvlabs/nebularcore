package event

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig defines retry behavior for event handlers
type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
}

// DefaultRetryConfig provides sensible default retry settings
var DefaultRetryConfig = RetryConfig{
	MaxRetries:  3,
	InitialWait: time.Second,
	MaxWait:     time.Second * 10,
}

// WithRetry wraps a handler with retry logic
func WithRetry(handler Handler, config RetryConfig) Handler {
	return func(ctx context.Context, event Message) error {
		var lastErr error
		wait := config.InitialWait

		for i := 0; i < config.MaxRetries; i++ {
			if err := handler(ctx, event); err != nil {
				lastErr = err

				// Log retry attempt
				event.Metadata.Set("retry_count", fmt.Sprintf("%d", i+1))
				event.Metadata.Set("last_error", err.Error())

				// Wait before retrying
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(wait):
					// Exponential backoff with max wait
					wait *= 2
					if wait > config.MaxWait {
						wait = config.MaxWait
					}
					continue
				}
			}
			return nil
		}

		return fmt.Errorf("max retries reached: %v", lastErr)
	}
}

// WithTimeout wraps a handler with a timeout
func WithTimeout(handler Handler, timeout time.Duration) Handler {
	return func(ctx context.Context, event Message) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		done := make(chan error, 1)
		go func() {
			done <- handler(ctx, event)
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// WithRecovery wraps a handler with panic recovery
func WithRecovery(handler Handler) Handler {
	return func(ctx context.Context, event Message) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("handler panic: %v", r)
			}
		}()
		return handler(ctx, event)
	}
}

// Chain combines multiple handlers into a single handler
func Chain(handlers ...Handler) Handler {
	return func(ctx context.Context, event Message) error {
		for _, h := range handlers {
			if err := h(ctx, event); err != nil {
				return err
			}
		}
		return nil
	}
}
