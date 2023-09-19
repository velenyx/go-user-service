package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

func NewClient(ctx context.Context, uri, username, password string) (db *mongo.Client, err error) {
	var mongoDBURL string

	if username == "" && password == "" {
		mongoDBURL = fmt.Sprintf("mongodb+srv://%s", uri)
	} else {
		mongoDBURL = fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority", username, password, uri)
	}

	log.Printf("Connecting to MongoDB with URL: %s", mongoDBURL)

	clientOptions := options.Client().ApplyURI(mongoDBURL)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to MongoDB due to error: %v", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("Failed to ping to MongoDB due to error: %v", err)
	}

	log.Println("Connected to MongoDB")

	return client, nil
}
