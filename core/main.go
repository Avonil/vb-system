package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"twitch-bot-system/core/pkg/core"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	// 🔥 Створюємо пул з'єднань, щоб Ядро знало, кому пересилати івенти
	activeClients = make(map[*websocket.Conn]string)
	clientsMu     sync.Mutex
)

// 🔥 Додали поле Client, щоб ловити назву бота з хендшейку
type WSMessage struct {
	Type   string          `json:"type"`
	Client string          `json:"client,omitempty"`
	Data   json.RawMessage `json:"data"`
}

type UpdatePayload struct {
	TwitchID string `json:"twitchId"`
	Username string `json:"username"`
}

func main() {
	_ = godotenv.Load()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017/veris?authSource=admin"
	}

	client, err := connectMongo(mongoURI)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✅ MongoDB Connected")

	redisURI := os.Getenv("REDIS_URI")
	if redisURI == "" {
		redisURI = "redis://localhost:6379/0" // Дефолтний локальний редіс
	}
	core.InitRedis(redisURI) // Викликаємо твою функцію з redis.go

	botManager := core.NewBotManager(client)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Реєструємо нового клієнта як "unknown"
		clientsMu.Lock()
		activeClients[conn] = "unknown"
		clientsMu.Unlock()

		// При відключенні видаляємо з пулу
		defer func() {
			clientsMu.Lock()
			delete(activeClients, conn)
			clientsMu.Unlock()
		}()

		log.Println("🔌 Client Connected (API or Bot)")

		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var wsMsg WSMessage
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				log.Println("JSON Error:", err)
				continue
			}

			switch wsMsg.Type {

			// 🔥 ЛОВИМО HANDSHAKE ВІД БОТА
			case "HANDSHAKE":
				clientsMu.Lock()
				activeClients[conn] = wsMsg.Client // Записуємо, що це "twitch-bot"
				clientsMu.Unlock()
				log.Printf("🤝 Handshake received from: %s", wsMsg.Client)

			case "API_HANDSHAKE":
				log.Println("🤝 API Handshake received")

			case "USER_UPDATED", "COMMANDS_UPDATED":
				var payload UpdatePayload
				if err := json.Unmarshal(wsMsg.Data, &payload); err != nil {
					continue
				}
				log.Printf("📥 Event: %s for %s", wsMsg.Type, payload.Username)

				go func() {
					botManager.ReloadUser(payload.TwitchID)

					// 🔥 НАЙГОЛОВНІШЕ: БРОДКАСТ!
					// Пересилаємо цей івент всім підключеним Твіч-ботам
					broadcastMsg, _ := json.Marshal(wsMsg)
					clientsMu.Lock()
					for clientConn, clientType := range activeClients {
						if clientType == "twitch-bot" {
							clientConn.WriteMessage(websocket.TextMessage, broadcastMsg)
							log.Printf("📤 Broadcasted %s to twitch-bot", wsMsg.Type)
						}
					}
					clientsMu.Unlock()
				}()

			case "STREAM_ONLINE":
				var streamData struct {
					TwitchID string `json:"twitchId"`
					Username string `json:"username"`
					Title    string `json:"title"`
					Category string `json:"category"`
				}
				if err := json.Unmarshal(wsMsg.Data, &streamData); err != nil {
					log.Printf("❌ Error parsing STREAM_ONLINE: %v", err)
					continue
				}

				log.Printf("🎥 Streamer %s is LIVE! Category: %s", streamData.Username, streamData.Category)

				// 1. Шукаємо стрімера в базі (або кеші)
				streamer, exists := botManager.GetUser(streamData.TwitchID)
				if !exists || streamer == nil {
					err := botManager.ReloadUser(streamData.TwitchID)
					if err != nil {
						log.Printf("⚠️ Cannot load streamer %s for alerts", streamData.TwitchID)
						continue
					}
					// Отримуємо стрімера знову, ігноруючи другий параметр через "_"
					streamer, _ = botManager.GetUser(streamData.TwitchID)
				}

				// 2. Перевіряємо, чи взагалі увімкнені алерти
				if !streamer.Modules.Alerts.Enabled {
					log.Printf("🔕 Alerts disabled for %s", streamData.Username)
					continue
				}

				// 3. Формуємо базовий пакет даних для ботів
				alertPayload := map[string]interface{}{
					"streamer": streamData.Username,
					"title":    streamData.Title,
					"category": streamData.Category,
					"url":      "https://twitch.tv/" + streamData.Username,
				}

				// 4. Відправляємо в Discord (через Redis)
				if streamer.Modules.Alerts.Discord.Enabled && streamer.Modules.Alerts.Discord.ChannelID != "" {
					discordPayload := alertPayload
					discordPayload["platform"] = "discord"
					discordPayload["channelId"] = streamer.Modules.Alerts.Discord.ChannelID
					discordPayload["message"] = streamer.Modules.Alerts.Discord.Message

					msgBytes, _ := json.Marshal(discordPayload)
					core.RedisClient.Publish(context.Background(), "alerts_discord", msgBytes)
					log.Printf("📤 Sent Discord alert for %s to Redis", streamData.Username)
				}

				// 5. Відправляємо в Telegram (через Redis)
				if streamer.Modules.Alerts.Telegram.Enabled && streamer.Modules.Alerts.Telegram.ChatID != "" {
					telegramPayload := alertPayload
					telegramPayload["platform"] = "telegram"
					telegramPayload["chatId"] = streamer.Modules.Alerts.Telegram.ChatID
					telegramPayload["message"] = streamer.Modules.Alerts.Telegram.Message

					msgBytes, _ := json.Marshal(telegramPayload)
					core.RedisClient.Publish(context.Background(), "alerts_telegram", msgBytes)
					log.Printf("📤 Sent Telegram alert for %s to Redis", streamData.Username)
				}

			case "CHAT_COMMAND":
				var cmdData struct {
					Channel string   `json:"channel"`
					User    string   `json:"user"`
					Command string   `json:"command"`
					Args    []string `json:"args"`
				}
				if err := json.Unmarshal(wsMsg.Data, &cmdData); err != nil {
					continue
				}

				log.Printf("🎮 [Twitch] Command: !%s from %s in [%s]", cmdData.Command, cmdData.User, cmdData.Channel)

				messages := core.ProcessCommand(client, botManager, "twitch", cmdData.Channel, cmdData.User, cmdData.Command, cmdData.Args)

				for _, msg := range messages {
					replyPayload := map[string]interface{}{
						"type": "SEND_MESSAGE",
						"data": map[string]string{
							"channel": cmdData.Channel,
							"text":    msg,
						},
					}
					replyBytes, _ := json.Marshal(replyPayload)
					conn.WriteMessage(websocket.TextMessage, replyBytes)
				}

			case "CLIENT_READY", "OBS_STREAM_STARTING":
				var payload struct {
					TwitchID string `json:"twitchId"`
				}
				if err := json.Unmarshal(wsMsg.Data, &payload); err == nil {
					botManager.ReloadUser(payload.TwitchID)
				}
			}
		}
		log.Println("🔌 Client Disconnected")
	})

	http.HandleFunc("/api/bot/command", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Platform string   `json:"platform"`
			ServerID string   `json:"serverId"`
			User     string   `json:"user"`
			Command  string   `json:"command"`
			Args     []string `json:"args"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		streamer, err := botManager.GetUserByPlatformID(req.Platform, req.ServerID)
		if err != nil || streamer == nil {
			http.Error(w, "Streamer not found", http.StatusNotFound)
			return
		}

		log.Printf("🎮 [%s] Command: !%s from %s (Streamer: %s)", req.Platform, req.Command, req.User, streamer.Username)

		messages := core.ProcessCommand(client, botManager, req.Platform, streamer.Username, req.User, req.Command, req.Args)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"responses": messages,
		})
	})

	port := os.Getenv("WS_PORT")
	if port == "" {
		port = "9000"
	}
	srv := &http.Server{Addr: ":" + port}

	go func() {
		log.Printf("🟢 Veris Core running on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("✅ Server exiting")
}

func connectMongo(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(uri)
	return mongo.Connect(ctx, clientOptions)
}
