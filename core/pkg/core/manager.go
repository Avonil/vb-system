package core

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"twitch-bot-system/core/pkg/db" // Імпорт твоїх моделей

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// BotManager тримає кеш юзерів
type BotManager struct {
	mu          sync.RWMutex
	activeUsers map[string]*db.User // Key: TwitchID
	mongoClient *mongo.Client
}

func NewBotManager(client *mongo.Client) *BotManager {
	return &BotManager{
		activeUsers: make(map[string]*db.User),
		mongoClient: client,
	}
}

// ReloadUser - Йде в базу і оновлює пам'ять
func (m *BotManager) ReloadUser(twitchID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Важливо: колекція "users", як в API
	collection := m.mongoClient.Database("veris").Collection("users")

	var user db.User
	// Шукаємо по twitchId (camelCase, як записав Mongoose)
	err := collection.FindOne(ctx, bson.M{"twitchId": twitchID}).Decode(&user)
	if err != nil {
		return fmt.Errorf("failed to fetch user %s: %v", twitchID, err)
	}

	m.mu.Lock()
	m.activeUsers[twitchID] = &user
	m.mu.Unlock()

	log.Printf("♻️ [RAM UPDATE] User: %s | Commands: %d | Token: ...%s",
		user.Username,
		len(user.Modules.Chat.CustomCommands),
		user.Auth.AccessToken[len(user.Auth.AccessToken)-5:], // показуємо останні 5 символів токена
	)

	return nil
}

// GetUser - дає дані з пам'яті (блискавично)
func (m *BotManager) GetUser(twitchID string) (*db.User, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, exists := m.activeUsers[twitchID]
	return user, exists
}

// RemoveUser - Видаляє юзера з кешу, щоб не жерти RAM
func (m *BotManager) RemoveUser(twitchID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.activeUsers[twitchID]; exists {
		delete(m.activeUsers, twitchID)
		log.Printf("🗑️ [RAM CLEANUP] User removed from memory: %s", twitchID)
	}
}

// GetUserByUsername шукає юзера в кеші по його Twitch нікнейму (назві каналу)
func (m *BotManager) GetUserByUsername(username string) *db.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, user := range m.activeUsers {
		// Порівнюємо без урахування регістру, бо Twitch іноді грається з великими літерами
		if strings.EqualFold(user.Username, username) {
			return user
		}
	}
	return nil
}

func (m *BotManager) ReloadUserByUsername(username string) (*db.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := m.mongoClient.Database("veris").Collection("users")
	var user db.User

	// Шукаємо без урахування регістру (case-insensitive) за допомогою Regex
	filter := bson.M{"username": bson.M{"$regex": "^" + username + "$", "$options": "i"}}
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("user %s not found in db: %v", username, err)
	}

	// Записуємо в кеш по TwitchID (бо це наш головний ключ)
	m.mu.Lock()
	m.activeUsers[user.TwitchID] = &user
	m.mu.Unlock()

	log.Printf("♻️ [LAZY LOAD] User loaded into RAM: %s", user.Username)
	return &user, nil
}

// GetUserByPlatformID шукає стрімера за ID сервера Discord або чату Telegram
func (m *BotManager) GetUserByPlatformID(platform, serverID string) (*db.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := m.mongoClient.Database("veris").Collection("users")
	var filter bson.M

	if platform == "discord" {
		filter = bson.M{"modules.alerts.discord.serverId": serverID}
	} else if platform == "telegram" {
		filter = bson.M{"modules.alerts.telegram.chatId": serverID}
	} else {
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}

	var user db.User
	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("streamer not found for %s server %s", platform, serverID)
	}

	// Кешуємо в оперативку
	m.mu.Lock()
	m.activeUsers[user.TwitchID] = &user
	m.mu.Unlock()

	return &user, nil
}
