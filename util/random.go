package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ============================================
// RANDOM STRING GENERATORS
// ============================================

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numberBytes = "0123456789"
)

// RandomString generates a random string of specified length
func RandomString(length int) string {
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		result[i] = letterBytes[num.Int64()]
	}
	return string(result)
}

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return min + n.Int64()
}

// RandomBool generates a random boolean
func RandomBool() bool {
	return RandomInt(0, 1) == 1
}

// ============================================
// USER DATA GENERATORS
// ============================================

// RandomUsername generates a random username
func RandomUsername() string {
	return fmt.Sprintf("user_%s", RandomString(5))
}

// RandomEmail generates a random email address
func RandomEmail() string {
	return fmt.Sprintf("%s@%s.com", RandomString(6), RandomString(4))
}

// RandomPassword generates a random password and returns both plaintext and hash
func RandomPassword() (plaintext string, hash string, err error) {
	plaintext = RandomString(12) + fmt.Sprintf("%d", RandomInt(100, 999))
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return plaintext, string(hashedBytes), nil
}

// RandomPublicKey generates a random public key (simulated)
func RandomPublicKey() string {
	// In production, this would be a real Ed25519/X25519 public key
	// For testing, we generate a random base64 string
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)
	return base64.StdEncoding.EncodeToString(keyBytes)
}

// RandomProfilePictureURL generates a random profile picture URL
func RandomProfilePictureURL() string {
	avatarServices := []string{
		"https://i.kratos.cc/300?img=%d",
		"https://kratos.me/api/portraits/men/%d.jpg",
		"https://kratos.me/api/portraits/women/%d.jpg",
	}
	service := avatarServices[RandomInt(0, int64(len(avatarServices)-1))]
	return fmt.Sprintf(service, RandomInt(1, 70))
}

// ============================================
// MESSAGE DATA GENERATORS
// ============================================

// RandomEncryptedContent generates random encrypted content (simulated)
func RandomEncryptedContent() string {
	// Simulate encrypted message - in production this would be real E2E encrypted data
	contentLength := RandomInt(50, 500)
	encryptedBytes := make([]byte, contentLength)
	rand.Read(encryptedBytes)
	return base64.StdEncoding.EncodeToString(encryptedBytes)
}

// RandomClientMessageID generates a random client message ID
func RandomClientMessageID() string {
	return fmt.Sprintf("msg_%d_%s", time.Now().UnixNano(), uuid.New().String()[:8])
}

// RandomMessageContent generates random plain text message for testing
func RandomMessageContent() string {
	messages := []string{
		"Hey, how are you?",
		"What's up?",
		"Did you see the game last night?",
		"Let's meet tomorrow at 3pm",
		"Thanks for your help!",
		"I'll be there in 10 minutes",
		"Can you send me that file?",
		"Great work on the project!",
		"See you later!",
		"Have a great day!",
		"This is awesome!",
		"I agree with you",
		"Let me think about it",
		"Sounds good to me",
		"I'm working on it now",
	}
	return messages[RandomInt(0, int64(len(messages)-1))]
}

// ============================================
// TIMESTAMP GENERATORS
// ============================================

// RandomPastTimestamp generates a random timestamp in the past
func RandomPastTimestamp(daysAgo int) time.Time {
	hoursAgo := RandomInt(0, int64(daysAgo*24))
	return time.Now().Add(-time.Hour * time.Duration(hoursAgo))
}

// RandomRecentTimestamp generates a timestamp within the last few hours
func RandomRecentTimestamp() time.Time {
	minutesAgo := RandomInt(0, 120) // Last 2 hours
	return time.Now().Add(-time.Minute * time.Duration(minutesAgo))
}

// ============================================
// STRUCT GENERATORS FOR SQLC PARAMS
// ============================================

// CreateUserParams represents parameters for creating a user
type CreateUserParams struct {
	Username          string
	Email             string
	PasswordHash      string
	PublicKey         string
	ProfilePictureURL *string
}

// RandomUserParams generates random parameters for creating a user
func RandomUserParams() CreateUserParams {
	_, passwordHash, _ := RandomPassword()
	profileURL := RandomProfilePictureURL()

	return CreateUserParams{
		Username:          RandomUsername(),
		Email:             RandomEmail(),
		PasswordHash:      passwordHash,
		PublicKey:         RandomPublicKey(),
		ProfilePictureURL: &profileURL,
	}
}

// CreateMessageParams represents parameters for creating a message
type CreateMessageParams struct {
	ConversationID   uuid.UUID
	SenderID         uuid.UUID
	EncryptedContent string
	ClientMessageID  *string
}

// RandomMessageParams generates random parameters for creating a message
func RandomMessageParams(conversationID, senderID uuid.UUID) CreateMessageParams {
	clientMsgID := RandomClientMessageID()

	return CreateMessageParams{
		ConversationID:   conversationID,
		SenderID:         senderID,
		EncryptedContent: RandomEncryptedContent(),
		ClientMessageID:  &clientMsgID,
	}
}

// ============================================
// HELPER FUNCTIONS FOR TESTING
// ============================================

// RandomUUID generates a random UUID
func RandomUUID() uuid.UUID {
	return uuid.New()
}

// RandomUUIDs generates multiple random UUIDs
func RandomUUIDs(count int) []uuid.UUID {
	uuids := make([]uuid.UUID, count)
	for i := 0; i < count; i++ {
		uuids[i] = uuid.New()
	}
	return uuids
}

// NullString returns a pointer to string (for nullable fields)
func NullString(s string) *string {
	return &s
}

// NullTime returns a pointer to time.Time (for nullable fields)
func NullTime(t time.Time) *time.Time {
	return &t
}

// ============================================
// BULK DATA GENERATORS
// ============================================

// GenerateRandomUsers generates N random users
func GenerateRandomUsers(n int) []CreateUserParams {
	users := make([]CreateUserParams, n)
	for i := 0; i < n; i++ {
		users[i] = RandomUserParams()
	}
	return users
}

// GenerateRandomMessages generates N random messages for a conversation
func GenerateRandomMessages(n int, conversationID uuid.UUID, senderIDs []uuid.UUID) []CreateMessageParams {
	if len(senderIDs) == 0 {
		panic("need at least one sender ID")
	}

	messages := make([]CreateMessageParams, n)
	for i := 0; i < n; i++ {
		senderID := senderIDs[RandomInt(0, int64(len(senderIDs)-1))]
		messages[i] = RandomMessageParams(conversationID, senderID)
	}
	return messages
}
