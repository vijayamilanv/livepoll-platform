package database

import (
	"context"
	"log"
	"time"

	"livepoll/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

// Connect establishes a MongoDB connection and creates required indexes.
func Connect() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(config.App.MongoURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("[mongo] connect error: %v", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("[mongo] ping error: %v", err)
	}

	Client = client
	DB = client.Database(config.App.MongoDB)
	log.Println("[mongo] connected to", config.App.MongoDB)

	ensureIndexes()
}

func ensureIndexes() {
	ctx := context.Background()

	// users: unique email index
	usersCol := DB.Collection("users")
	_, err := usersCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("[mongo] users email index: %v", err)
	}

	// polls: unique shareCode index
	pollsCol := DB.Collection("polls")
	_, err = pollsCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "shareCode", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("[mongo] polls shareCode index: %v", err)
	}

	// votes: compound unique index (pollId + voterFingerprint) — one vote per device per poll
	votesCol := DB.Collection("votes")
	_, err = votesCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "pollId", Value: 1}, {Key: "voterFingerprint", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("[mongo] votes compound index: %v", err)
	}

	log.Println("[mongo] indexes ensured")
}

// Col is a convenience helper to get a collection.
func Col(name string) *mongo.Collection {
	return DB.Collection(name)
}
