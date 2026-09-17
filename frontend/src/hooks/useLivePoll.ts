import { useEffect, useRef, useCallback, useState } from 'react';
import type { VoteCount } from '../api/polls';

interface WSMessage {
  pollId: string;
  counts: VoteCount[];
}

const WS_BASE = import.meta.env.VITE_WS_URL || 'ws://localhost:8080';
const RECONNECT_DELAY_MS = 3000;

export const useLivePoll = (shareCode: string, pollId: string | null) => {
  const [counts, setCounts] = useState<VoteCount[]>([]);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const shouldReconnect = useRef(true);

  const connect = useCallback(() => {
    if (!pollId) return;

    const url = `${WS_BASE}/ws/poll/${shareCode}`;
    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
    };

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data);
        if (msg.counts) {
          setCounts(msg.counts);
        }
      } catch {
        // ignore malformed messages
      }
    };

    ws.onclose = () => {
      setConnected(false);
      if (shouldReconnect.current) {
        reconnectTimer.current = setTimeout(connect, RECONNECT_DELAY_MS);
      }
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [shareCode, pollId]);

  useEffect(() => {
    shouldReconnect.current = true;
    connect();

    return () => {
      shouldReconnect.current = false;
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current);
      wsRef.current?.close();
    };
  }, [connect]);

  return { counts, connected };
};
