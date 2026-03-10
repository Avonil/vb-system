package db

import (
	"time"
)

// User - це дзеркало твого Mongoose Schema
type User struct {
	ID          string `bson:"_id"`      // Mongo ID
	TwitchID    string `bson:"twitchId"` // 🔥 Головний ключ
	Username    string `bson:"username"`
	DisplayName string `bson:"displayName"`
	AvatarURL   string `bson:"avatarUrl"`

	IsStreamer bool `bson:"isStreamer"`
	IsAdmin    bool `bson:"isAdmin"`

	Auth    AuthData     `bson:"auth"`
	Config  GlobalConfig `bson:"config"`
	Modules Modules      `bson:"modules"`

	CreatedAt time.Time `bson:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt"`
}

// ViewerState - зберігає кастомні змінні глядача для конкретного стрімера (економіка, RPG тощо)
type ViewerState struct {
	ID         string                 `bson:"_id,omitempty"`
	StreamerID string                 `bson:"streamerId"` // У кого на стрімі
	ViewerID   string                 `bson:"viewerId"`   // Хто глядач
	ViewerName string                 `bson:"viewerName"`
	Variables  map[string]interface{} `bson:"variables"` // Тут будуть "coins": 100, "mana": 50
	UpdatedAt  time.Time              `bson:"updatedAt"`
}

// VisualCommand - структура команди, зібраної нодами в Tauri
type VisualCommand struct {
	Trigger string                   `bson:"trigger"` // наприклад "roulette"
	Logic   []map[string]interface{} `bson:"logic"`   // Масив кроків (JSON), які Ядро має виконати
}

type AuthData struct {
	AccessToken  string    `bson:"accessToken"`
	RefreshToken string    `bson:"refreshToken"`
	Scopes       []string  `bson:"scopes"`
	ExpiresAt    time.Time `bson:"expiresAt"`
}

type GlobalConfig struct {
	Language   string `bson:"language"`
	Prefix     string `bson:"prefix"`
	BotEnabled bool   `bson:"botEnabled"`
	Timezone   string `bson:"timezone"`
}

// --- MODULES ---

type Modules struct {
	Chat       ChatModule       `bson:"chat"`
	Moderation ModerationModule `bson:"moderation"` // якщо вона в тебе є
	Alerts     AlertsModule     `bson:"alerts"`     // 🔥 Додаємо модуль сповіщень
}

// AlertsModule - Головний модуль сповіщень
type AlertsModule struct {
	Enabled  bool          `bson:"enabled"`
	Discord  DiscordAlert  `bson:"discord"`
	Telegram TelegramAlert `bson:"telegram"`
}

// DiscordAlert - Налаштування для Discord бота
type DiscordAlert struct {
	Enabled   bool   `bson:"enabled"`
	ServerID  string `bson:"serverId"`  // ID гільдії (сервера)
	ChannelID string `bson:"channelId"` // ID каналу для сповіщень
	Message   string `bson:"message"`   // Текст сповіщення (напр. "Гей, @everyone, я онлайн!")
}

// TelegramAlert - Налаштування для Telegram бота
type TelegramAlert struct {
	Enabled bool   `bson:"enabled"`
	ChatID  string `bson:"chatId"`  // ID групи або каналу в ТГ
	Message string `bson:"message"` // Текст сповіщення
}

// Додаємо в структуру ChatModule масив візуальних команд
type ChatModule struct {
	Enabled        bool            `bson:"enabled"`
	AnnounceLive   bool            `bson:"announceLive"`
	CustomCommands []CustomCommand `bson:"customCommands"`
	VisualCommands []VisualCommand `bson:"visualCommands"` // 🔥 Сюди Tauri буде зберігати JSON
}

type CustomCommand struct {
	Trigger   string `bson:"trigger"`
	Response  string `bson:"response"`
	Cooldown  int    `bson:"cooldown"`
	UserLevel string `bson:"userLevel"`
	Enabled   bool   `bson:"enabled"`
}

type ModerationModule struct {
	Enabled     bool     `bson:"enabled"`
	BannedWords []string `bson:"bannedWords"`
}
