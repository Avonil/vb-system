package core

import (
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

// ProcessCommand - Універсальний обробник команд для Twitch, Discord та Telegram
func ProcessCommand(client *mongo.Client, botManager *BotManager, platform, channel, user, command string, args []string) []string {
	var responses []string

	// 1. Знаходимо стрімера (Lazy Load)
	streamer := botManager.GetUserByUsername(channel)
	if streamer == nil {
		var err error
		streamer, err = botManager.ReloadUserByUsername(channel)
		if err != nil {
			log.Printf("❌ Cannot load streamer %s: %v", channel, err)
			return responses // Порожній масив
		}
	}
	// Очищаємо вхідну команду від зайвих знаків оклику та пробілів
	cleanCommand := strings.TrimSpace(strings.TrimPrefix(command, "!"))

	// 2. ПЕРЕВІРКА: Візуальні команди
	visualCommandFound := false
	for _, vc := range streamer.Modules.Chat.VisualCommands {
		// Очищаємо тригер з бази від знаку оклику
		cleanTrigger := strings.TrimSpace(strings.TrimPrefix(vc.Trigger, "!"))

		if strings.EqualFold(cleanTrigger, cleanCommand) {
			visualCommandFound = true
			log.Printf("✨ [Visual] Executing script for command: %s", cleanCommand)
			responses = ExecuteVisualScript(client, streamer, user, vc.Logic)
			break
		}
	}

	if visualCommandFound {
		return responses
	}

	// 3. ПЕРЕВІРКА: Звичайні текстові команди
	for _, cmd := range streamer.Modules.Chat.CustomCommands {
		// Зверни увагу: тут має бути правильна назва поля з твоєї структури db.User
		// Наприклад: cmd.Command, cmd.Trigger чи cmd.Name
		// (Якщо в тебе поле називається інакше, заміни cmd.Command на твоє)
		dbTrigger := strings.TrimSpace(strings.TrimPrefix(cmd.Trigger, "!"))

		log.Printf("🔍 Comparing received '%s' with DB '%s'", cleanCommand, dbTrigger)

		if strings.EqualFold(dbTrigger, cleanCommand) {
			log.Printf("✅ Match found! Replying: %s", cmd.Response)

			// Замінюємо змінні (якщо є)
			reply := strings.ReplaceAll(cmd.Response, "{user}", user)
			responses = append(responses, reply)
			break
		}
	}

	return responses
}
