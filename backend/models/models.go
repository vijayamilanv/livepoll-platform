package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a registered user.
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}

// PollOption is a single answer choice inside a poll.
type PollOption struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

// Poll is the core document stored in MongoDB.
type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ShareCode string             `bson:"shareCode" json:"shareCode"`
	Question  string             `bson:"question" json:"question"`
	Options   []PollOption       `bson:"options" json:"options"`
	CreatedBy primitive.ObjectID `bson:"createdBy" json:"createdBy"`
	IsActive  bool               `bson:"isActive" json:"isActive"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	ClosesAt  *time.Time         `bson:"closesAt,omitempty" json:"closesAt,omitempty"`
}

// Vote represents a single vote cast by an audience member.
type Vote struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID          primitive.ObjectID `bson:"pollId" json:"pollId"`
	OptionID        string             `bson:"optionId" json:"optionId"`
	VoterFingerprint string            `bson:"voterFingerprint" json:"voterFingerprint"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
}

// VoteCount holds the live count for a single option (returned to clients).
type VoteCount struct {
	OptionID string `json:"optionId"`
	Count    int64  `json:"count"`
}

// PollWithCounts bundles a poll with its live vote counts.
type PollWithCounts struct {
	Poll
	Counts []VoteCount `json:"counts"`
}
