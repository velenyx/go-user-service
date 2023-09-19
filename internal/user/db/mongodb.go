package db

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"some/internal/user"
	"some/pkg/logging"
)

type db struct {
	collection *mongo.Collection
	logger     *logging.Logger
}

func (d *db) Create(ctx context.Context, user *user.User) (string, error) {
	d.logger.Debug("create user")
	result, err := d.collection.InsertOne(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed to create user due to error: %v", err)
	}
	d.logger.Debug("convert InsertedID to ObjectID")
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if ok {
		return oid.Hex(), nil
	}
	d.logger.Trace(user)
	return "", fmt.Errorf("failed to convert objectid to hex, probably oid: %s", oid)
}

func (d *db) FindAll(ctx context.Context) (users []*user.User, err error) {
	cursor, err := d.collection.Find(ctx, bson.M{})
	if err != nil {
		// TODO ErrorEntityNotFound
		if errors.Is(err, mongo.ErrNoDocuments) {
			return users, fmt.Errorf("not found")
		}
		return users, fmt.Errorf("failed to find all users due to error: %v", err)
	}
	if err = cursor.All(ctx, &users); err != nil {
		return users, fmt.Errorf("failed to decode users from cursor due to error: %v", err)
	}
	return users, nil
}

func (d *db) FindOne(ctx context.Context, id string) (u *user.User, err error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return u, fmt.Errorf("failed to convert hex to objectid due to error, hex: %v", err)
	}
	filter := bson.M{"_id": oid}

	result := d.collection.FindOne(ctx, filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			// TODO ErrEntityNotFound
			return u, fmt.Errorf("not found")
		}
		return u, fmt.Errorf("failed to find user by id: %s due to error: %v", id, result.Err())
	}
	if err = result.Decode(&u); err != nil {
		return u, fmt.Errorf("failed to decode user(id: %s) from DB due to error: %v", id, err)
	}
	return u, nil
}

func (d *db) Update(ctx context.Context, user *user.User) error {
	oid, err := primitive.ObjectIDFromHex(user.ID)
	if err != nil {
		return fmt.Errorf("failed to convert hex to objectid due to error, hex: %v", err)
	}

	filter := bson.M{"_id": oid}

	userBytes, err := bson.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user due to error: %v", err)
	}

	var updateUserObj bson.M
	err = bson.Unmarshal(userBytes, &updateUserObj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal user bytes due to error: %v", err)
	}

	delete(updateUserObj, "_id")

	update := bson.M{
		"$set": updateUserObj,
	}

	result, err := d.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user(id: %s) due to error: %v", user.ID, err)
	}
	if result.MatchedCount == 0 {
		// TODO ErrorEntityNotFound
		return fmt.Errorf("user(id: %s) not found", user.ID)
	}

	d.logger.Tracef("Matched %d documents and updated %d documents", result.MatchedCount, result.ModifiedCount)

	return nil
}

func (d *db) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("failed to convert hex to objectid due to error, hex: %v", err)
	}

	filter := bson.M{"_id": oid}

	result, err := d.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete user(id: %s) due to error: %v", id, err)
	}
	if result.DeletedCount == 0 {
		// TODO ErrorEntityNotFound
		return fmt.Errorf("user(id: %s) not found", id)
	}

	d.logger.Tracef("Deleted %d documents", result.DeletedCount)

	return nil
}

func NewStorage(database *mongo.Database, collection string, logger *logging.Logger) user.Storage {

	return &db{
		collection: database.Collection(collection),
		logger:     logger,
	}
}
