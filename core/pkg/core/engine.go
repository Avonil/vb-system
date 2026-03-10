package core

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"twitch-bot-system/core/pkg/db"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ExecuteVisualScript проходить по масиву JSON-нодів (які стрімер зібрав у Tauri) і виконує їх
func ExecuteVisualScript(client *mongo.Client, streamer *db.User, viewer string, logic []map[string]interface{}) []string {
	var messagesToSend []string

	// 1. Отримуємо стейт юзера з БД (його коїни, ману і т.д.)
	state, err := GetViewerState(client, streamer.TwitchID, viewer, viewer)
	if err != nil {
		log.Println("Error getting viewer state:", err)
		return nil
	}

	// Хелпер: безпечно дістаємо число з БД (Mongo може повернути int32, int64 або float64)
	getVar := func(name string) float64 {
		if val, ok := state.Variables[name]; ok {
			switch v := val.(type) {
			case float64:
				return v
			case int32:
				return float64(v)
			case int64:
				return float64(v)
			case int:
				return float64(v)
			}
		}
		return 0 // Якщо змінної ще немає, вона = 0
	}

	// 2. Головний цикл по нодах
	for _, node := range logic {
		action, ok := node["action"].(string)
		if !ok {
			continue
		}

		switch action {
		// НОДА 1: Перевірка умови (наприклад, чи є 100 коїнів для ставки)
		case "check_balance":
			varName := node["var"].(string)
			condition := node["condition"].(string)
			reqValue := node["value"].(float64) // З JSON числа завжди приходять як float64

			currentVal := getVar(varName)
			passed := false

			switch condition {
			case ">=":
				passed = currentVal >= reqValue
			case "<=":
				passed = currentVal <= reqValue
			case "==":
				passed = currentVal == reqValue
			}

			if !passed {
				// Якщо перевірка провалена (немає грошей) - генеруємо повідомлення про помилку і ЗУПИНЯЄМО скрипт
				if failMsg, ok := node["fail_msg"].(string); ok {
					msg := strings.ReplaceAll(failMsg, "{user}", "@"+viewer)
					messagesToSend = append(messagesToSend, msg)
				}
				return messagesToSend
			}

		// НОДА 2: Математика (додати, відняти, встановити)
		case "math":
			varName := node["var"].(string)
			operation := node["operation"].(string)
			val := node["value"].(float64)

			currentVal := getVar(varName)
			newVal := currentVal

			switch operation {
			case "+":
				newVal += val
			case "-":
				newVal -= val
			case "set":
				newVal = val
			}

			// Оновлюємо значення локально в пам'яті і одразу в БД
			state.Variables[varName] = newVal
			UpdateViewerVariable(client, streamer.TwitchID, viewer, varName, newVal)

		// НОДА 3: Відправка повідомлення в чат
		case "send_chat":
			if text, ok := node["text"].(string); ok {
				// Динамічна заміна. Наприклад "Твій баланс: {coins}"
				text = strings.ReplaceAll(text, "{user}", "@"+viewer)
				for k, v := range state.Variables {
					placeholder := fmt.Sprintf("{%s}", k)
					text = strings.ReplaceAll(text, placeholder, fmt.Sprintf("%v", v))
				}
				messagesToSend = append(messagesToSend, text)
			}

		// НОДА 4: Рандомна ймовірність (наприклад, рулетка 50/50)
		case "random_chance":
			chance := int(node["win_chance"].(float64))
			roll := rand.Intn(100) + 1 // Від 1 до 100

			if roll <= chance {
				// Виграш - запускаємо вкладений блок нодів (рекурсія!)
				if winLogic, ok := node["win"].([]interface{}); ok {
					messagesToSend = append(messagesToSend, ExecuteVisualScript(client, streamer, viewer, convertLogic(winLogic))...)
				}
			} else {
				// Програш
				if loseLogic, ok := node["lose"].([]interface{}); ok {
					messagesToSend = append(messagesToSend, ExecuteVisualScript(client, streamer, viewer, convertLogic(loseLogic))...)
				}
			}
		}
	}

	return messagesToSend
}

// Хелпер для конвертації вкладених JSON-масивів
func convertLogic(raw []interface{}) []map[string]interface{} {
	var res []map[string]interface{}
	for _, item := range raw {
		if m, ok := item.(map[string]interface{}); ok {
			res = append(res, m)
		}
	}
	return res
}
