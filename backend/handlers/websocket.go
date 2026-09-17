package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"livepoll/hub"
	"livepoll/redisdb"
	"livepoll/services"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		// Origin check handled by CORS middleware; allow all for WS in dev.
		// In production lock this to config.App.FrontendURL.
		return true
	},
	HandshakeTimeout: 10 * time.Second,
}

type countItem struct {
	OptionID string `json:"optionId"`
	Count    int64  `json:"count"`
}

type wsSnapshot struct {
	PollID string      `json:"pollId"`
	Counts []countItem `json:"counts"`
}

// WSPoll handles GET /ws/poll/:shareCode — upgrades to WebSocket.
func WSPoll(c *gin.Context) {
	shareCode := c.Param("shareCode")

	// Fetch poll to get its internal ID and verify it exists
	poll, err := services.GetPollByShareCode(shareCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	pollID := poll.ID.Hex()

	// Build initial snapshot from Redis (warmed by GetPollByShareCode)
	initialSnap := buildInitialSnapshot(pollID)

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// upgrader already wrote the error response
		return
	}

	// Get or create room and join (blocks until client disconnects)
	room := hub.GetOrCreate(pollID)
	room.Join(conn, initialSnap)
}

// buildInitialSnapshot returns a JSON snapshot of current counts from Redis.
func buildInitialSnapshot(pollID string) []byte {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	countsMap, _ := redisdb.GetCounts(ctx, pollID)
	items := make([]countItem, 0, len(countsMap))
	for k, v := range countsMap {
		n, _ := strconv.ParseInt(v, 10, 64)
		items = append(items, countItem{OptionID: k, Count: n})
	}

	snap := wsSnapshot{PollID: pollID, Counts: items}
	b, err := json.Marshal(snap)
	if err != nil {
		return []byte(fmt.Sprintf(`{"pollId":%q,"counts":[]}`, pollID))
	}
	return b
}
