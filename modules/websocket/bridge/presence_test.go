package bridge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/volvlabs/nebularcore/modules/event"
	eventmocks "github.com/volvlabs/nebularcore/modules/event/mocks"
)

func TestPresenceTopic(t *testing.T) {
	if got, want := PresenceTopic("abc-123"), "presence.user.abc-123"; got != want {
		t.Fatalf("PresenceTopic() = %q, want %q", got, want)
	}
}

func TestPresenceBroadcaster_OnlinePublishesImmediately(t *testing.T) {
	bus := new(eventmocks.Bus)
	published := make(chan struct{}, 1)
	bus.On("Publish", mock.Anything, mock.MatchedBy(func(m event.Message) bool {
		return m.EventType == "presence.user.u1"
	})).Run(func(mock.Arguments) { published <- struct{}{} }).Return(nil)

	pb := NewPresenceBroadcaster(bus, 50*time.Millisecond)
	pb.OnPresenceChange("u1", true)

	select {
	case <-published:
	case <-time.After(time.Second):
		t.Fatal("expected immediate publish on online transition")
	}
	bus.AssertNumberOfCalls(t, "Publish", 1)
}

func TestPresenceBroadcaster_OfflineIsDebounced(t *testing.T) {
	bus := new(eventmocks.Bus)
	published := make(chan struct{}, 1)
	bus.On("Publish", mock.Anything, mock.Anything).Run(func(mock.Arguments) { published <- struct{}{} }).Return(nil)

	pb := NewPresenceBroadcaster(bus, 100*time.Millisecond)
	pb.OnPresenceChange("u1", false)

	select {
	case <-published:
		t.Fatal("offline publish should not fire before debounce window elapses")
	case <-time.After(30 * time.Millisecond):
	}

	select {
	case <-published:
	case <-time.After(time.Second):
		t.Fatal("expected debounced publish after window elapses")
	}
	bus.AssertNumberOfCalls(t, "Publish", 1)
}

func TestPresenceBroadcaster_ReconnectCancelsDebounce(t *testing.T) {
	bus := new(eventmocks.Bus)
	published := make(chan event.Message, 2)
	bus.On("Publish", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		published <- args.Get(1).(event.Message)
	}).Return(nil)

	pb := NewPresenceBroadcaster(bus, 100*time.Millisecond)
	pb.OnPresenceChange("u1", false) // starts debounce timer
	pb.OnPresenceChange("u1", true)  // cancels it, publishes online immediately

	select {
	case msg := <-published:
		if msg.EventType != "presence.user.u1" {
			t.Fatalf("unexpected event type %q", msg.EventType)
		}
	case <-time.After(time.Second):
		t.Fatal("expected online publish")
	}

	// Wait past the original debounce window to confirm the offline publish
	// never fires.
	select {
	case msg := <-published:
		t.Fatalf("unexpected second publish after reconnect: %+v", msg)
	case <-time.After(200 * time.Millisecond):
	}
	bus.AssertNumberOfCalls(t, "Publish", 1)
}
