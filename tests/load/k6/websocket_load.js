import ws from 'k6/ws';
import { check, sleep } from 'k6';
import { Counter, Trend, Rate } from 'k6/metrics';

// Custom metrics
const wsConnections = new Counter('ws_connections');
const wsMessages = new Counter('ws_messages_sent');
const wsErrors = new Counter('ws_errors');
const wsLatency = new Trend('ws_latency_ms');
const wsConnectRate = new Rate('ws_connect_success');

// Test configuration — ramp to target concurrent WebSocket connections.
export const options = {
  scenarios: {
    // Ramp-up scenario: gradually increase to target connections
    ramp_up: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 1000 },   // Warm up
        { duration: '1m',  target: 10000 },   // Ramp to 10k
        { duration: '2m',  target: 50000 },   // Ramp to 50k
        { duration: '3m',  target: 100000 },  // Ramp to 100k
        { duration: '5m',  target: 100000 },  // Hold at 100k
        { duration: '1m',  target: 0 },       // Ramp down
      ],
    },
  },
  thresholds: {
    'ws_connect_success': ['rate>0.95'],       // 95%+ connections succeed
    'ws_latency_ms':      ['p(95)<500'],        // p95 latency < 500ms
    'ws_errors':          ['count<1000'],        // Less than 1000 errors
  },
};

const WS_URL = __ENV.WS_URL || 'ws://server:8080/ws';

export default function () {
  const startTime = new Date().getTime();

  const res = ws.connect(WS_URL, {}, function (socket) {
    wsConnections.add(1);
    wsConnectRate.add(1);

    const connectLatency = new Date().getTime() - startTime;
    wsLatency.add(connectLatency);

    socket.on('open', function () {
      // Subscribe to a topic.
      const topic = `load.test.user${__VU}`;
      socket.send(JSON.stringify({
        id: `sub-${__VU}-${__ITER}`,
        type: 'subscribe',
        topic: topic,
      }));
      wsMessages.add(1);
    });

    socket.on('message', function (msg) {
      const parsed = JSON.parse(msg);

      // On subscription confirmation, send a ping.
      if (parsed.type === 'subscribed') {
        socket.send(JSON.stringify({
          id: `ping-${__VU}-${__ITER}`,
          type: 'ping',
        }));
        wsMessages.add(1);
      }

      // On pong, send a publish.
      if (parsed.type === 'pong') {
        const sendTime = new Date().getTime();
        socket.send(JSON.stringify({
          id: `pub-${__VU}-${__ITER}`,
          type: 'publish',
          topic: `load.test.user${__VU}`,
          payload: { ts: sendTime, vu: __VU },
        }));
        wsMessages.add(1);
      }
    });

    socket.on('error', function (e) {
      wsErrors.add(1);
      wsConnectRate.add(0);
    });

    // Hold the connection open for the duration of the test iteration.
    // k6 keeps the socket alive for the duration specified.
    socket.setTimeout(function () {
      socket.close();
    }, 60000); // Hold for 60 seconds.
  });

  check(res, {
    'status is 101 (switching protocols)': (r) => r && r.status === 101,
  });

  // Small sleep between iterations to prevent thundering herd.
  sleep(Math.random() * 2);
}
