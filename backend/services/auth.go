package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"livepoll/config"
	"livepoll/database"
	"livepoll/middleware"
	"livepoll/models"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailTaken     = errors.New("email already registered")
	ErrInvalidCreds   = errors.New("invalid email or password")
)

// Signup creates a new user and returns a signed JWT.
func Signup(email, password string) (string, *models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", nil, err
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := database.Col("users").InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return "", nil, ErrEmailTaken
		}
		return "", nil, err
	}

	user.ID = res.InsertedID.(primitive.ObjectID)
	token, err := generateToken(user)
	return token, user, err
}

// Login verifies credentials and returns a signed JWT.
func Login(email, password string) (string, *models.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := database.Col("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return "", nil, ErrInvalidCreds
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", nil, ErrInvalidCreds
	}

	token, err := generateToken(&user)
	return token, &user, err
}

func generateToken(user *models.User) (string, error) {
	claims := middleware.Claims{
		UserID: user.ID.Hex(),
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.App.JWTSecret))
}
