package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateToken(userID int64) string {
	data := fmt.Sprintf("%d-%s-%d", userID, GenerateUUID(), time.Now().UnixNano())
	return HashString(data)
}

func HashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func HashPassword(password string) string {
	return HashString(password + "game_backend_salt")
}

func VerifyPassword(password, hashedPassword string) bool {
	return HashPassword(password) == hashedPassword
}

func GenerateRandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GenerateRoomID() string {
	return fmt.Sprintf("ROOM_%s", GenerateRandomString(10))
}

func GetUserTokenKey(userID int64) string {
	return fmt.Sprintf("user:token:%d", userID)
}

func GetOnlineUserKey() string {
	return "online:users"
}
