package hub

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"livepoll/redisdb"

	"github.com/gorilla/websocket"
)

// client represents a single WebSocket connection inside a room.
type client struct {
	conn *websocket.Conn
	send chan []byte
}

// Room is a goroutine-safe set of WebSocket clients watching a single poll.
type Room struct {
	pollID  string
	mu      sync.RWMutex
	clients map[*client]struct{}
}

// Hub manages all active poll rooms and routes Redis pub/sub events to them.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*Room // keyed by pollID
}

// Global singleton hub.
var global = &Hub{rooms: make(map[string]*Room)}

// GetOrCreate returns the Room for pollID, creating it (and subscribing to Redis) if needed.
func GetOrCreate(pollID string) *Room {
	global.mu.RLock()
	r, ok := global.rooms[pollID]
	global.mu.RUnlock()
	if ok {
		return r
	}

	global.mu.Lock()
	defer global.mu.Unlock()
	// Double-check after acquiring write lock
	if r, ok = global.rooms[pollID]; ok {
		return r
	}

	r = &Room{pollID: pollID, clients: make(map[*client]struct{})}
	global.rooms[pollID] = r

	// Start the Redis subscriber goroutine for this room
	go r.subscribeRedis()

	return r
}

// Join adds a WebSocket connection to this room and starts its write pump.
// It sends the current snapshot (initialMsg) immediately on connect.
func (r *Room) Join(conn *websocket.Conn, initialMsg []byte) {
	c := &client{conn: conn, send: make(chan []byte, 64)}

	r.mu.Lock()
	r.clients[c] = struct{}{}
	r.mu.Unlock()

	// Send the snapshot first
	if initialMsg != nil {
		c.send <- initialMsg
	}

	// Start write pump — this goroutine owns the conn write side
	go c.writePump(r)
	// Start read pump (just drains pings / close frames, no app-level reads needed)
	c.readPump(r)
}

// Broadcast sends a message to all connected clients in this room.
func (r *Room) Broadcast(msg []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for c := range r.clients {
		select {
		case c.send <- msg:
		default:
			// Slow client — drop the message rather than blocking
		}
	}
}

// remove unregisters a client from this room.
func (r *Room) remove(c *client) {
	r.mu.Lock()
	delete(r.clients, c)
	r.mu.Unlock()
	close(c.send)
}

// subscribeRedis listens on the Redis pub/sub channel for this poll and
// broadcasts every received message to all WebSocket clients in this room.
// This goroutine lives for the lifetime of the room.
func (r *Room) subscribeRedis() {
	ctx := context.Background()
	ps := redisdb.Subscribe(ctx, r.pollID)
	defer ps.Close()

	ch := ps.Channel()
	log.Printf("[hub] subscribed to Redis channel for poll %s", r.pollID)

	for msg := range ch {
		r.Broadcast([]byte(msg.Payload))
	}
	log.Printf("[hub] Redis subscription ended for poll %s", r.pollID)
}

// writePump drains the client's send channel and writes to the WebSocket.
func (c *client) writePump(r *Room) {
	defer func() {
		c.conn.Close()
	}()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}

// readPump reads from the WebSocket (only used to detect disconnects / handle pings).
func (c *client) readPump(r *Room) {
	defer func() {
		r.remove(c)
		c.conn.Close()
	}()
	c.conn.SetReadLimit(512)
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// CountsSnapshot is the JSON shape sent to a client when they first connect.
type CountsSnapshot struct {
	PollID string              `json:"pollId"`
	Counts []countEntry        `json:"counts"`
}

type countEntry struct {
	OptionID string `json:"optionId"`
	Count    int64  `json:"count"`
}

// MarshalSnapshot converts a raw counts payload to JSON bytes.
func MarshalSnapshot(pollID string, counts map[string]int64) []byte {
	entries := make([]countEntry, 0, len(counts))
	for k, v := range counts {
		entries = append(entries, countEntry{OptionID: k, Count: v})
	}
	snap := CountsSnapshot{PollID: pollID, Counts: entries}
	b, _ := json.Marshal(snap)
	return b
}
