package core

import (
	"context"
	"encoding/json"
	"log"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var ctx = context.Background()

func InitRedis(uri string) {
	opts, err := redis.ParseURL(uri)
	if err != nil {
		log.Fatalf("❌ Redis URL Error: %v", err)
	}

	RedisClient = redis.NewClient(opts)
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Redis Connection Error: %v", err)
	}
	log.Println("✅ Redis Connected")
}

// Структура для 4-х параметрів стріму
type StreamStartPayload struct {
	TwitchID string `json:"twitchId"`
	Username string `json:"username"`
	Title    string `json:"title"`
	Category string `json:"category"`
}

// PublishStreamAlert - відправляє сигнал в канал "stream_alerts"
func PublishStreamAlert(payload StreamStartPayload) {
	data, _ := json.Marshal(payload)
	
	// Публікуємо в канал. DS і TG боти мають слухати цей канал!
	err := RedisClient.Publish(ctx, "stream_alerts", data).Err()
	if err != nil {
		log.Printf("❌ Failed to publish stream alert: %v", err)
	} else {
		log.Printf("📣 [REDIS] Stream alert published for %s", payload.Username)
	}
}