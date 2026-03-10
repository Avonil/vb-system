package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Streamer - модель для збереження токенів
type Streamer struct {
	Username     string    `bson:"username"`
	AccessToken  string    `bson:"accessToken"`  // Зверни увагу на CamelCase, як в API
	RefreshToken string    `bson:"refreshToken"`
	UpdatedAt    time.Time `bson:"updatedAt"`
}

var Client *mongo.Client
var StreamerCollection *mongo.Collection

func InitDB(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	Client = client
	// Використовуємо ту ж базу "veris", але окрему колекцію для секретів стрімерів
	// АБО можемо писати в ту ж "users", просто оновлюючи поля.
	// Давай поки окремо "streamer_auth", щоб не поламати логіку паспортів.
	StreamerCollection = client.Database("veris").Collection("streamer_auth")
	
	log.Println("✅ MongoDB Connected!")
	return nil
}

// UpsertStreamer - оновлює або створює запис про токени
func UpsertStreamer(username, token, refresh string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"username": username}
	update := bson.M{
		"$set": bson.M{
			"username":     username,
			"accessToken":  token,
			"refreshToken": refresh,
			"updatedAt":    time.Now(),
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := StreamerCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetStreamer - отримує токени
func GetStreamer(username string) (*Streamer, error) {
	var s Streamer
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := StreamerCollection.FindOne(ctx, bson.M{"username": username}).Decode(&s)
	return &s, err
}

// GetAllStreamers - для перепідключення ботів при старті
func GetAllStreamers() ([]Streamer, error) {
	var streamers []Streamer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := StreamerCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &streamers); err != nil {
		return nil, err
	}
	return streamers, nil
}