package bridge

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/volvlabs/nebularcore/modules/event"
	"github.com/volvlabs/nebularcore/modules/websocket/connections"
)

// PresenceTopic builds the exact WebSocket topic a client subscribes to in
// order to receive presence updates for a specific user, e.g.
// "presence.user.abc-123".
func PresenceTopic(userID string) string {
	return fmt.Sprintf("presence.user.%s", userID)
}

// presencePayload is the payload delivered on a user's presence topic.
type presencePayload struct {
	UserID     string     `json:"userId"`
	Online     bool       `json:"online"`
	LastSeenAt *time.Time `json:"lastSeenAt"`
}

// PresenceBroadcaster implements connections.PresenceListener. It turns raw
// 0<->1 connection-count transitions into presence events on the event bus:
// "came online" is published immediately, "went offline" is debounced by a
// grace window so a quick reconnect (page reload, brief network blip) never
// produces a visible online/offline flicker.
type PresenceBroadcaster struct {
	bus      event.Bus
	debounce time.Duration

	// offlineTimers holds a pending debounced "went offline" publish per
	// userID, so a reconnect within the window can cancel it.
	offlineTimers sync.Map
}

// NewPresenceBroadcaster creates a PresenceBroadcaster. debounce is the grace
// window before a "went offline" transition is actually published; a value
// <= 0 disables debouncing (offline is published immediately, same as
// online).
func NewPresenceBroadcaster(bus event.Bus, debounce time.Duration) *PresenceBroadcaster {
	return &PresenceBroadcaster{
		bus:      bus,
		debounce: debounce,
	}
}

// OnPresenceChange implements connections.PresenceListener.
func (p *PresenceBroadcaster) OnPresenceChange(userID string, online bool) {
	if online {
		if t, ok := p.offlineTimers.LoadAndDelete(userID); ok {
			t.(*time.Timer).Stop()
		}
		p.publish(userID, true, nil)
		return
	}

	if p.debounce <= 0 {
		p.publish(userID, false, timePtr(time.Now().UTC()))
		return
	}

	timer := time.AfterFunc(p.debounce, func() {
		p.offlineTimers.Delete(userID)
		p.publish(userID, false, timePtr(time.Now().UTC()))
	})

	if existing, loaded := p.offlineTimers.LoadOrStore(userID, timer); loaded {
		// Shouldn't normally happen (an "online" transition always clears
		// the pending timer first), but guard against a stray double-fire.
		timer.Stop()
		_ = existing
	}
}

func (p *PresenceBroadcaster) publish(userID string, online bool, lastSeenAt *time.Time) {
	topic := PresenceTopic(userID)

	msg, err := event.NewMessage(topic, "websocket-presence", presencePayload{
		UserID:     userID,
		Online:     online,
		LastSeenAt: lastSeenAt,
	})
	if err != nil {
		log.Err(err).Str("user_id", userID).Msg("presence: failed to create event")
		return
	}

	if err := p.bus.Publish(context.Background(), msg); err != nil {
		log.Err(err).Str("user_id", userID).Str("topic", topic).Msg("presence: failed to publish event")
	}
}

var _ connections.PresenceListener = (*PresenceBroadcaster)(nil)

func timePtr(t time.Time) *time.Time { return &t }
