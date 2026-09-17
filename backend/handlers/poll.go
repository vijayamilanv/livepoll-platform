package handlers

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"

	"livepoll/middleware"
	"livepoll/services"

	"github.com/gin-gonic/gin"
)

type createPollRequest struct {
	Question string   `json:"question" binding:"required"`
	Options  []string `json:"options" binding:"required"`
}

type voteRequest struct {
	OptionID string `json:"optionId" binding:"required"`
}

// CreatePoll handles POST /polls (auth required)
func CreatePoll(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createPollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	poll, err := services.CreatePoll(req.Question, req.Options, claims.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, poll)
}

// GetPoll handles GET /polls/:shareCode (public)
func GetPoll(c *gin.Context) {
	shareCode := c.Param("shareCode")
	poll, err := services.GetPollByShareCode(shareCode)
	if err != nil {
		if errors.Is(err, services.ErrPollNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch poll"})
		return
	}
	c.JSON(http.StatusOK, poll)
}

// GetMyPolls handles GET /polls/mine (auth required)
func GetMyPolls(c *gin.Context) {
	claims := middleware.GetClaims(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	polls, err := services.GetMyPolls(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch polls"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"polls": polls})
}

// Vote handles POST /polls/:shareCode/vote (public)
func Vote(c *gin.Context) {
	shareCode := c.Param("shareCode")

	var req voteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fingerprint := voterFingerprint(c)
	counts, err := services.CastVote(shareCode, req.OptionID, fingerprint)
	if err != nil {
		switch {
		case errors.Is(err, services.ErrPollNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		case errors.Is(err, services.ErrInvalidOption):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option"})
		case errors.Is(err, services.ErrAlreadyVoted):
			c.JSON(http.StatusConflict, gin.H{"error": "you have already voted on this poll"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not cast vote"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"counts": counts})
}

// voterFingerprint creates a SHA-256 hash of IP + User-Agent to identify voters
// without requiring login. This is a best-effort deduplication — a determined
// user with VPN+private mode can bypass it. Decision documented in README.
func voterFingerprint(c *gin.Context) string {
	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")
	raw := fmt.Sprintf("%s|%s", ip, ua)
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}
