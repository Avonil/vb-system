package core

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"twitch-bot-system/core/pkg/db" // Твій шлях до моделей
)

// GetViewerState - дістає інфу про глядача (або створює пусту, якщо він вперше пише команду)
func GetViewerState(client *mongo.Client, streamerID, viewerID, viewerName string) (*db.ViewerState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("veris").Collection("viewer_states")

	var state db.ViewerState
	filter := bson.M{"streamerId": streamerID, "viewerId": viewerID}

	err := collection.FindOne(ctx, filter).Decode(&state)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Якщо глядач вперше на стрімі юзає RPG-команду - створюємо йому дефолтний профіль
			newState := &db.ViewerState{
				StreamerID: streamerID,
				ViewerID:   viewerID,
				ViewerName: viewerName,
				Variables:  map[string]interface{}{"coins": 0, "xp": 0}, // Базові значення
				UpdatedAt:  time.Now(),
			}
			_, insertErr := collection.InsertOne(ctx, newState)
			return newState, insertErr
		}
		return nil, err
	}

	return &state, nil
}

// UpdateViewerVariable - точково оновлює якусь змінну (наприклад, додає 100 коїнів)
func UpdateViewerVariable(client *mongo.Client, streamerID, viewerID, varName string, value interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("veris").Collection("viewer_states")
	filter := bson.M{"streamerId": streamerID, "viewerId": viewerID}
	
	// Оновлюємо конкретне поле в об'єкті Variables
	update := bson.M{
		"$set": bson.M{
			fmt.Sprintf("variables.%s", varName): value,
			"updatedAt": time.Now(),
		},
	}
	// Якщо юзера ще не було, створюємо (upsert)
	opts := options.Update().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}