package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"livepoll/database"
	"livepoll/models"
	"livepoll/redisdb"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrPollNotFound    = errors.New("poll not found")
	ErrInvalidOption   = errors.New("invalid option id for this poll")
	ErrAlreadyVoted    = errors.New("you have already voted on this poll")
)

const (
	maxQuestionLen = 500
	maxOptions     = 10
	minOptions     = 2
	maxOptionLen   = 200
)

// CreatePoll validates inputs and inserts a new poll into MongoDB.
func CreatePoll(question string, optionTexts []string, userID string) (*models.Poll, error) {
	// Server-side validation (mirrors frontend rules)
	if len(question) == 0 || len(question) > maxQuestionLen {
		return nil, fmt.Errorf("question must be 1–%d characters", maxQuestionLen)
	}
	if len(optionTexts) < minOptions {
		return nil, fmt.Errorf("poll must have at least %d options", minOptions)
	}
	if len(optionTexts) > maxOptions {
		return nil, fmt.Errorf("poll may have at most %d options", maxOptions)
	}

	options := make([]models.PollOption, 0, len(optionTexts))
	for i, text := range optionTexts {
		if len(text) == 0 || len(text) > maxOptionLen {
			return nil, fmt.Errorf("option %d must be 1–%d characters", i+1, maxOptionLen)
		}
		options = append(options, models.PollOption{
			ID:   uuid.New().String(),
			Text: text,
		})
	}

	creatorID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	poll := &models.Poll{
		ShareCode: generateShareCode(),
		Question:  question,
		Options:   options,
		CreatedBy: creatorID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := database.Col("polls").InsertOne(ctx, poll)
	if err != nil {
		return nil, err
	}
	poll.ID = res.InsertedID.(primitive.ObjectID)
	return poll, nil
}

// GetPollByShareCode fetches a poll and its live counts (Redis-first, Mongo fallback).
func GetPollByShareCode(shareCode string) (*models.PollWithCounts, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var poll models.Poll
	err := database.Col("polls").FindOne(ctx, bson.M{"shareCode": shareCode}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, err
	}

	counts, err := fetchOrWarmCounts(ctx, &poll)
	if err != nil {
		return nil, err
	}

	return &models.PollWithCounts{Poll: poll, Counts: counts}, nil
}

// GetMyPolls returns all polls created by a user (newest first).
func GetMyPolls(userID string) ([]models.PollWithCounts, error) {
	creatorID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.Col("polls").Find(ctx,
		bson.M{"createdBy": creatorID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var polls []models.Poll
	if err = cursor.All(ctx, &polls); err != nil {
		return nil, err
	}

	result := make([]models.PollWithCounts, 0, len(polls))
	for _, p := range polls {
		p := p
		counts, _ := fetchOrWarmCounts(ctx, &p)
		result = append(result, models.PollWithCounts{Poll: p, Counts: counts})
	}
	return result, nil
}

// CastVote writes a vote to Mongo, increments Redis, and publishes the update.
// Returns updated counts for immediate HTTP response.
func CastVote(shareCode, optionID, voterFingerprint string) ([]models.VoteCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Fetch poll to validate optionID belongs to it
	var poll models.Poll
	err := database.Col("polls").FindOne(ctx, bson.M{"shareCode": shareCode}).Decode(&poll)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPollNotFound
		}
		return nil, err
	}

	// Validate that the requested optionID belongs to this poll
	validOption := false
	for _, opt := range poll.Options {
		if opt.ID == optionID {
			validOption = true
			break
		}
	}
	if !validOption {
		return nil, ErrInvalidOption
	}

	// Write vote to MongoDB (unique index will reject duplicates)
	vote := models.Vote{
		PollID:           poll.ID,
		OptionID:         optionID,
		VoterFingerprint: voterFingerprint,
		CreatedAt:        time.Now().UTC(),
	}
	_, err = database.Col("votes").InsertOne(ctx, vote)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, ErrAlreadyVoted
		}
		return nil, err
	}

	// Increment Redis counter — O(1), non-blocking for the DB
	if _, err = redisdb.IncrVote(ctx, poll.ID.Hex(), optionID); err != nil {
		// Non-fatal: log and continue; counts will be rebuilt from Mongo next request
		fmt.Printf("[vote] redis HINCRBY failed: %v\n", err)
	}

	// Fetch all counts from Redis for the pub/sub payload + response
	countsMap, _ := redisdb.GetCounts(ctx, poll.ID.Hex())
	counts := mapToCounts(poll.Options, countsMap)

	// Publish event so all WebSocket subscribers get the update instantly
	payload := buildCountsPayload(poll.ID.Hex(), counts)
	_ = redisdb.Publish(ctx, poll.ID.Hex(), payload)

	return counts, nil
}

// fetchOrWarmCounts gets counts from Redis; on cache miss rebuilds from Mongo.
func fetchOrWarmCounts(ctx context.Context, poll *models.Poll) ([]models.VoteCount, error) {
	countsMap, err := redisdb.GetCounts(ctx, poll.ID.Hex())
	if err != nil {
		return nil, err
	}

	if countsMap == nil {
		// Cold cache: aggregate from MongoDB and warm Redis
		countsMap, err = aggregateVotesFromMongo(ctx, poll.ID)
		if err != nil {
			return nil, err
		}
		// Warm Redis
		warmData := make(map[string]int64, len(countsMap))
		for k, v := range countsMap {
			n, _ := strconv.ParseInt(v, 10, 64)
			warmData[k] = n
		}
		if len(warmData) > 0 {
			_ = redisdb.SetCounts(ctx, poll.ID.Hex(), warmData)
		}
	}

	return mapToCounts(poll.Options, countsMap), nil
}

// aggregateVotesFromMongo counts votes per option directly from the votes collection.
func aggregateVotesFromMongo(ctx context.Context, pollID primitive.ObjectID) (map[string]string, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"pollId": pollID}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$optionId"},
			{Key: "count", Value: bson.M{"$sum": 1}},
		}}},
	}
	cursor, err := database.Col("votes").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := map[string]string{}
	for cursor.Next(ctx) {
		var row struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err = cursor.Decode(&row); err == nil {
			result[row.ID] = strconv.FormatInt(row.Count, 10)
		}
	}
	return result, nil
}

// mapToCounts converts a Redis HGETALL map to the []VoteCount response type.
func mapToCounts(options []models.PollOption, countsMap map[string]string) []models.VoteCount {
	counts := make([]models.VoteCount, 0, len(options))
	for _, opt := range options {
		var count int64
		if v, ok := countsMap[opt.ID]; ok {
			count, _ = strconv.ParseInt(v, 10, 64)
		}
		counts = append(counts, models.VoteCount{OptionID: opt.ID, Count: count})
	}
	return counts
}

// buildCountsPayload serialises counts into JSON for the Redis pub/sub message.
func buildCountsPayload(pollID string, counts []models.VoteCount) string {
	s := fmt.Sprintf(`{"pollId":%q,"counts":[`, pollID)
	for i, c := range counts {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprintf(`{"optionId":%q,"count":%d}`, c.OptionID, c.Count)
	}
	s += "]}"
	return s
}

// generateShareCode creates a URL-safe 8-character random code.
func generateShareCode() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:8]
}
