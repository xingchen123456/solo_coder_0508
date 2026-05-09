package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"sync"
	"time"
)

type Generator struct {
	mu        sync.Mutex
	counter   int64
	timestamp int64
	nodeID    int64
}

var defaultGenerator *Generator

func init() {
	defaultGenerator = New(1)
}

func New(nodeID int64) *Generator {
	if nodeID < 0 || nodeID > 1023 {
		nodeID = 1
	}
	return &Generator{
		nodeID:    nodeID,
		timestamp: time.Now().UnixMilli(),
	}
}

func (g *Generator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < g.timestamp {
		now = g.timestamp
	}

	if now == g.timestamp {
		g.counter++
	} else {
		g.timestamp = now
		g.counter = 0
	}

	id := (now << 20) | (g.nodeID << 10) | g.counter
	return strconv.FormatInt(id, 10)
}

func (g *Generator) GenerateHex() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < g.timestamp {
		now = g.timestamp
	}

	if now == g.timestamp {
		g.counter++
	} else {
		g.timestamp = now
		g.counter = 0
	}

	id := (now << 20) | (g.nodeID << 10) | g.counter

	bytes := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		bytes[i] = byte(id & 0xFF)
		id >>= 8
	}

	return hex.EncodeToString(bytes)
}

func Generate() string {
	return defaultGenerator.Generate()
}

func GenerateHex() string {
	return defaultGenerator.GenerateHex()
}

func GenerateShort() string {
	now := time.Now().Format("20060102150405")
	randomBytes := make([]byte, 4)
	_, _ = rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)
	return "P" + now + randomStr
}
